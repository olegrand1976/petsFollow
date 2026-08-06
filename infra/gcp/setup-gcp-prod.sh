#!/usr/bin/env bash
# Bootstrap GCP petsFollow PRODUCTION (SQL dédié, secrets *-prod, bucket médias).
# Usage: ./infra/gcp/setup-gcp-prod.sh
#
# Ne touche PAS à premedica-db-staging ni aux secrets petsfollow-database-url (staging).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env-prod.sh
source "${SCRIPT_DIR}/lib/gcp-env-prod.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

echo "=== petsFollow setup GCP PROD — ${GCP_PROJECT_ID} ==="
echo "  SQL instance : ${SQL_INSTANCE}"
echo "  Cloud Run    : ${API_SERVICE} / ${FRONTEND_SERVICE}"
echo "  Domaines     : ${CUSTOM_DOMAIN} + ${API_CUSTOM_DOMAIN}"
echo "  Bucket       : gs://${GCS_MEDIA_BUCKET}"

if [[ "$SQL_INSTANCE" == "premedica-db-staging" ]] || [[ "$CLOUDSQL_INSTANCE" == *premedica-db-staging* ]]; then
  echo "ERREUR: refuse d'utiliser l'instance staging en prod" >&2
  exit 1
fi

ensure_secret() {
  local name="$1"
  if ! gcloud secrets describe "$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "→ CREATE secret ${name}"
    gcloud secrets create "$name" --replication-policy=automatic --project="$GCP_PROJECT_ID" --quiet
  fi
}

add_secret_version() {
  local name="$1"
  local value="$2"
  ensure_secret "$name"
  echo -n "$value" | gcloud secrets versions add "$name" --data-file=- --project="$GCP_PROJECT_ID" --quiet
}

SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
CB_SA="$(gcloud projects describe "$GCP_PROJECT_ID" --format='value(projectNumber)')@cloudbuild.gserviceaccount.com"

if ! gcloud iam service-accounts describe "$SA_EMAIL" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ CREATE service account ${SERVICE_ACCOUNT}"
  gcloud iam service-accounts create "$SERVICE_ACCOUNT" \
    --display-name="petsFollow Cloud Run" --project="$GCP_PROJECT_ID" --quiet
fi

for role in roles/cloudsql.client roles/secretmanager.secretAccessor roles/run.invoker; do
  gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" --role="$role" --quiet >/dev/null 2>&1 || true
done
gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --member="serviceAccount:${CB_SA}" --role="roles/iam.serviceAccountUser" --quiet >/dev/null 2>&1 || true

echo "→ Cloud SQL instance ${SQL_INSTANCE}"
if ! gcloud sql instances describe "$SQL_INSTANCE" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ CREATE instance ${SQL_INSTANCE} (db-g1-small, POSTGRES_15, ~10–15 min)…"
  gcloud sql instances create "$SQL_INSTANCE" \
    --project="$GCP_PROJECT_ID" \
    --database-version=POSTGRES_15 \
    --tier=db-g1-small \
    --region="$GCP_RUN_REGION" \
    --storage-size=10 \
    --storage-auto-increase \
    --availability-type=ZONAL \
    --backup-start-time=03:00 \
    --maintenance-window-day=SUN \
    --maintenance-window-hour=4 \
    --quiet
else
  echo "  Instance ${SQL_INSTANCE} existe"
fi

# Backups on (prod) — best-effort si déjà créé sans.
gcloud sql instances patch "$SQL_INSTANCE" \
  --project="$GCP_PROJECT_ID" \
  --backup-start-time=03:00 \
  --quiet 2>/dev/null || true

echo "→ Database ${DB_NAME}"
if ! gcloud sql databases describe "$DB_NAME" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud sql databases create "$DB_NAME" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --quiet
else
  echo "  Base ${DB_NAME} existe"
fi

if ! gcloud sql users list --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --format='value(name)' | grep -qx "$DB_USER"; then
  DB_PASS="${PETSFOLLOW_PROD_DB_PASSWORD:-$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)}"
  echo "→ CREATE USER ${DB_USER}"
  gcloud sql users create "$DB_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$DB_PASS" --quiet
  add_secret_version "$PF_SM_DB_PASSWORD" "$DB_PASS"
else
  DB_PASS="$(gcloud secrets versions access latest --secret="$PF_SM_DB_PASSWORD" --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$DB_PASS" ]]; then
    echo "ERREUR: utilisateur ${DB_USER} existe mais secret ${PF_SM_DB_PASSWORD} vide" >&2
    exit 1
  fi
  echo "  Utilisateur ${DB_USER} existe"
fi

ENC_PASS="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$DB_PASS")"
DATABASE_URL="postgres://${DB_USER}:${ENC_PASS}@/${DB_NAME}?host=/cloudsql/${CLOUDSQL_INSTANCE}"
add_secret_version "$PF_SM_DATABASE_URL" "$DATABASE_URL"
echo "→ ${PF_SM_DATABASE_URL} mis à jour"

if ! gcloud sql users list --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --format='value(name)' | grep -qx "$MIGRATE_USER"; then
  MIGRATE_PASS="${PETSFOLLOW_PROD_MIGRATE_DB_PASSWORD:-$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)}"
  echo "→ CREATE USER ${MIGRATE_USER}"
  gcloud sql users create "$MIGRATE_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$MIGRATE_PASS" --quiet
  add_secret_version "$PF_SM_MIGRATE_DB_PASSWORD" "$MIGRATE_PASS"
else
  MIGRATE_PASS="$(gcloud secrets versions access latest --secret="$PF_SM_MIGRATE_DB_PASSWORD" --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$MIGRATE_PASS" ]]; then
    MIGRATE_PASS="$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)"
    echo "→ RESET secret migrate (user ${MIGRATE_USER} existait sans secret)"
    gcloud sql users set-password "$MIGRATE_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$MIGRATE_PASS" --quiet
    add_secret_version "$PF_SM_MIGRATE_DB_PASSWORD" "$MIGRATE_PASS"
  else
    echo "  Utilisateur ${MIGRATE_USER} existe"
  fi
fi

ENC_MIGRATE_PASS="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$MIGRATE_PASS")"
MIGRATE_DATABASE_URL="postgres://${MIGRATE_USER}:${ENC_MIGRATE_PASS}@/${DB_NAME}?host=/cloudsql/${CLOUDSQL_INSTANCE}"
add_secret_version "$PF_SM_MIGRATE_DATABASE_URL" "$MIGRATE_DATABASE_URL"
echo "→ ${PF_SM_MIGRATE_DATABASE_URL} mis à jour"

if ! gcloud secrets versions list "$PF_SM_JWT_SIGNING_KEY" --project="$GCP_PROJECT_ID" --limit=1 --format='value(name)' 2>/dev/null | grep -q .; then
  JWT_KEY="$(openssl rand -base64 48 | tr -d '/+=' | head -c 64)"
  add_secret_version "$PF_SM_JWT_SIGNING_KEY" "$JWT_KEY"
  echo "→ ${PF_SM_JWT_SIGNING_KEY} créé"
else
  echo "  ${PF_SM_JWT_SIGNING_KEY} existe"
fi

# Redis DB 15 (même host que staging /14) — l’app isole aussi via REDIS_KEY_PREFIX.
REDIS_HOST="$(python3 -c "from urllib.parse import urlparse; import sys
u=urlparse(sys.argv[1]) if sys.argv[1] else None
print((u.hostname if u else '') or '')
" "$(gcloud secrets versions access latest --secret=petsfollow-redis-url --project="$GCP_PROJECT_ID" 2>/dev/null || true)")"
REDIS_HOST="${REDIS_HOST:-10.200.0.2}"
PROD_REDIS_URL="redis://${REDIS_HOST}:6379/${REDIS_DB}"
if ! gcloud secrets versions list "$PF_SM_REDIS_URL" --project="$GCP_PROJECT_ID" --limit=1 --format='value(name)' 2>/dev/null | grep -q .; then
  add_secret_version "$PF_SM_REDIS_URL" "$PROD_REDIS_URL"
  echo "→ ${PF_SM_REDIS_URL} = ${PROD_REDIS_URL}"
else
  echo "  ${PF_SM_REDIS_URL} existe"
fi

COMPUTE_SA="$(gcloud projects describe "$GCP_PROJECT_ID" --format='value(projectNumber)')-compute@developer.gserviceaccount.com"
for secret in \
  "$PF_SM_DATABASE_URL" \
  "$PF_SM_MIGRATE_DATABASE_URL" \
  "$PF_SM_JWT_SIGNING_KEY" \
  "$PF_SM_REDIS_URL" \
  "$PF_SM_DB_PASSWORD" \
  "$PF_SM_MIGRATE_DB_PASSWORD" \
  petsfollow-smtp-password \
  petsfollow-gemini-api-key \
  petsfollow-retention-secret \
  petsfollow-auth-health-secret
do
  if gcloud secrets describe "$secret" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    gcloud secrets add-iam-policy-binding "$secret" \
      --project="$GCP_PROJECT_ID" \
      --member="serviceAccount:${SA_EMAIL}" \
      --role=roles/secretmanager.secretAccessor \
      --quiet >/dev/null 2>&1 || true
    gcloud secrets add-iam-policy-binding "$secret" \
      --project="$GCP_PROJECT_ID" \
      --member="serviceAccount:${COMPUTE_SA}" \
      --role=roles/secretmanager.secretAccessor \
      --quiet >/dev/null 2>&1 || true
  fi
done

TMP_SQL="$(mktemp)"
GCS_URI="gs://premedica-prod-2025-db-exports/admin/pg-petsfollow-prod-bootstrap-$(date +%Y%m%d%H%M%S).sql"
if gcloud storage buckets describe gs://premedica-prod-2025-db-exports --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ Grants bootstrap migrate/app (prod)"
  SQL_SA="$(gcloud sql instances describe "$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --format='value(serviceAccountEmailAddress)')"
  if [[ -n "$SQL_SA" ]]; then
    gcloud storage buckets add-iam-policy-binding gs://premedica-prod-2025-db-exports \
      --member="serviceAccount:${SQL_SA}" \
      --role="roles/storage.objectViewer" \
      --project="$GCP_PROJECT_ID" --quiet >/dev/null 2>&1 || true
  fi
  # pg_trgm requis migrations 000081+
  cat >"$TMP_SQL" <<EOSQL
CREATE EXTENSION IF NOT EXISTS pg_trgm;
GRANT CONNECT ON DATABASE ${DB_NAME} TO ${MIGRATE_USER};
GRANT CONNECT ON DATABASE ${DB_NAME} TO ${DB_USER};
GRANT CREATE ON DATABASE ${DB_NAME} TO ${MIGRATE_USER};
GRANT ALL ON SCHEMA public TO ${MIGRATE_USER};
GRANT USAGE ON SCHEMA public TO ${DB_USER};
EOSQL
  gcloud storage cp "$TMP_SQL" "$GCS_URI" --quiet
  gcloud sql import sql "$SQL_INSTANCE" "$GCS_URI" \
    --database="$DB_NAME" --project="$GCP_PROJECT_ID" --quiet || true
else
  echo "  Bucket premedica-db-exports absent — grants manuels si besoin"
fi
rm -f "$TMP_SQL"

echo "→ Bucket GCS médias prod"
GCS_MEDIA_BUCKET="$GCS_MEDIA_BUCKET" bash "${SCRIPT_DIR}/setup-gcs-media.sh"

cat <<EOF

OK — bootstrap prod terminé.

Prochaines étapes :
  1. OVH zone petsfollow.app — enregistrements A (voir doc 10 § Production / checklist OVH)
  2. make gcp-deploy-prod
  3. PETSFOLLOW_GCP_ENV=prod make gcp-domain   # ou make gcp-domain-prod
  4. Attendre certificat ACTIVE puis smoke :
       curl -fsS https://api.petsfollow.app/health
  5. Flutter : make play-android-bundle-prod

Secrets créés (ne pas confondre avec staging) :
  ${PF_SM_DATABASE_URL}
  ${PF_SM_MIGRATE_DATABASE_URL}
  ${PF_SM_JWT_SIGNING_KEY}
  ${PF_SM_REDIS_URL}

EOF
