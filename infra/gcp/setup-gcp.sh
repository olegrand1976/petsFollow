#!/usr/bin/env bash
# Bootstrap GCP petsFollow on premedica-prod-2025 (DB, secrets, SA, Artifact Registry).
# Usage: ./infra/gcp/setup-gcp.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

echo "=== petsFollow setup GCP — ${GCP_PROJECT_ID} ==="

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

echo "→ IAM Cloud Build"
for role in roles/run.admin roles/artifactregistry.writer roles/iam.serviceAccountUser roles/secretmanager.secretAccessor roles/cloudsql.admin; do
  gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
    --member="serviceAccount:${CB_SA}" --role="$role" --quiet >/dev/null 2>&1 || true
done

if ! gcloud iam service-accounts describe "$SA_EMAIL" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ CREATE service account ${SERVICE_ACCOUNT}"
  gcloud iam service-accounts create "$SERVICE_ACCOUNT" \
    --display-name="petsFollow Cloud Run" --project="$GCP_PROJECT_ID" --quiet
fi

gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
  --member="serviceAccount:${CB_SA}" --role="roles/iam.serviceAccountUser" --quiet >/dev/null 2>&1 || true

for role in roles/cloudsql.client roles/secretmanager.secretAccessor roles/run.invoker; do
  gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" --role="$role" --quiet >/dev/null 2>&1 || true
done

if ! gcloud artifacts repositories describe "$AR_REPO" --location="$GCP_AR_REGION" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ CREATE Artifact Registry ${AR_REPO}"
  gcloud artifacts repositories create "$AR_REPO" \
    --repository-format=docker --location="$GCP_AR_REGION" --project="$GCP_PROJECT_ID" --quiet
else
  echo "  Artifact Registry ${AR_REPO} existe"
fi

echo "→ PostgreSQL (${SQL_INSTANCE})"
if ! gcloud sql databases describe "$DB_NAME" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  gcloud sql databases create "$DB_NAME" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --quiet
else
  echo "  Base ${DB_NAME} existe"
fi

if ! gcloud sql users list --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --format='value(name)' | grep -qx "$DB_USER"; then
  DB_PASS="${PETSFOLLOW_DB_PASSWORD:-$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)}"
  echo "→ CREATE USER ${DB_USER}"
  gcloud sql users create "$DB_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$DB_PASS" --quiet
  add_secret_version "petsfollow-db-password" "$DB_PASS"
else
  DB_PASS="$(gcloud secrets versions access latest --secret=petsfollow-db-password --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$DB_PASS" ]]; then
    echo "ERREUR: utilisateur ${DB_USER} existe mais secret petsfollow-db-password vide" >&2
    exit 1
  fi
  echo "  Utilisateur ${DB_USER} existe"
fi

ENC_PASS="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$DB_PASS")"
DATABASE_URL="postgres://${DB_USER}:${ENC_PASS}@/${DB_NAME}?host=/cloudsql/${CLOUDSQL_INSTANCE}"
add_secret_version "petsfollow-database-url" "$DATABASE_URL"
echo "→ petsfollow-database-url mis à jour (runtime ${DB_USER})"

if ! gcloud sql users list --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --format='value(name)' | grep -qx "$MIGRATE_USER"; then
  MIGRATE_PASS="${PETSFOLLOW_MIGRATE_DB_PASSWORD:-$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)}"
  echo "→ CREATE USER ${MIGRATE_USER}"
  gcloud sql users create "$MIGRATE_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$MIGRATE_PASS" --quiet
  add_secret_version "petsfollow-migrate-db-password" "$MIGRATE_PASS"
else
  MIGRATE_PASS="$(gcloud secrets versions access latest --secret=petsfollow-migrate-db-password --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$MIGRATE_PASS" ]]; then
    MIGRATE_PASS="$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)"
    echo "→ RESET secret migrate (user ${MIGRATE_USER} existait sans secret)"
    gcloud sql users set-password "$MIGRATE_USER" --instance="$SQL_INSTANCE" --project="$GCP_PROJECT_ID" --password="$MIGRATE_PASS" --quiet
    add_secret_version "petsfollow-migrate-db-password" "$MIGRATE_PASS"
  else
    echo "  Utilisateur ${MIGRATE_USER} existe"
  fi
fi

ENC_MIGRATE_PASS="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$MIGRATE_PASS")"
MIGRATE_DATABASE_URL="postgres://${MIGRATE_USER}:${ENC_MIGRATE_PASS}@/${DB_NAME}?host=/cloudsql/${CLOUDSQL_INSTANCE}"
add_secret_version "petsfollow-migrate-database-url" "$MIGRATE_DATABASE_URL"
echo "→ petsfollow-migrate-database-url mis à jour (${MIGRATE_USER})"

if ! gcloud secrets versions list petsfollow-jwt-signing-key --project="$GCP_PROJECT_ID" --limit=1 --format='value(name)' 2>/dev/null | grep -q .; then
  JWT_KEY="$(openssl rand -base64 48 | tr -d '/+=' | head -c 64)"
  add_secret_version "petsfollow-jwt-signing-key" "$JWT_KEY"
  echo "→ petsfollow-jwt-signing-key créé"
else
  echo "  petsfollow-jwt-signing-key existe"
fi

COMPUTE_SA="$(gcloud projects describe "$GCP_PROJECT_ID" --format='value(projectNumber)')-compute@developer.gserviceaccount.com"
for secret in petsfollow-database-url petsfollow-migrate-database-url petsfollow-jwt-signing-key petsfollow-redis-url petsfollow-db-password petsfollow-migrate-db-password petsfollow-gemini-api-key; do
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

# Privilèges initiaux migrate (avant premières tables)
TMP_SQL="$(mktemp)"
cat >"$TMP_SQL" <<EOSQL
GRANT CONNECT ON DATABASE ${DB_NAME} TO ${MIGRATE_USER};
GRANT CONNECT ON DATABASE ${DB_NAME} TO ${DB_USER};
GRANT CREATE ON DATABASE ${DB_NAME} TO ${MIGRATE_USER};
GRANT ALL ON SCHEMA public TO ${MIGRATE_USER};
GRANT USAGE ON SCHEMA public TO ${DB_USER};
EOSQL
GCS_URI="gs://premedica-prod-2025-db-exports/admin/pg-petsfollow-bootstrap-$(date +%Y%m%d%H%M%S).sql"
if gcloud storage buckets describe gs://premedica-prod-2025-db-exports --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "→ Grants bootstrap migrate/app"
  gcloud storage cp "$TMP_SQL" "$GCS_URI" --quiet
  gcloud sql import sql "$SQL_INSTANCE" "$GCS_URI" \
    --database="$DB_NAME" --project="$GCP_PROJECT_ID" --quiet || true
else
  echo "  Bucket premedica-db-exports absent — grants via shared-postgres plus tard"
fi
rm -f "$TMP_SQL"

echo "→ Bucket GCS médias"
bash "${SCRIPT_DIR}/setup-gcs-media.sh"

# Storage objectAdmin déjà posé par setup-gcs-media ; storage.objectAdmin aussi au niveau projet optionnel
gcloud projects add-iam-policy-binding "$GCP_PROJECT_ID" \
  --member="serviceAccount:${SA_EMAIL}" --role="roles/storage.objectAdmin" --quiet >/dev/null 2>&1 || true

if [[ -d "${INFRA_ROOT}/shared-redis" ]]; then
  echo "→ Sync secrets Redis (infra partagée)"
  bash "${INFRA_ROOT}/shared-redis/setup-gcp.sh" 2>&1 | tail -12 || true
fi

if [[ -d "${INFRA_ROOT}/shared-postgres" ]]; then
  echo "→ Privilèges PostgreSQL petsfollow (idempotent)"
  bash "${INFRA_ROOT}/shared-postgres/setup-db-protection.sh" 2>&1 | tail -8 || true
fi

cat <<EOF

Prêt pour :
  ./infra/gcp/setup-github-deploy.sh
  gcloud builds submit --config=infra/gcp/cloudbuild.yaml --project=${GCP_PROJECT_ID}
  ./infra/gcp/postdeploy.sh --seed
  ./infra/gcp/setup-custom-domain.sh

EOF
