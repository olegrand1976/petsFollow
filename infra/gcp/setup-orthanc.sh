#!/usr/bin/env bash
# Déploie Orthanc PACS (Cloud Run scale-to-zero) + bucket DICOM + DB PostgreSQL.
# Usage:
#   ./infra/gcp/setup-orthanc.sh [IMAGE_URI]
#   ./infra/gcp/setup-orthanc.sh [IMAGE_URI] --deploy-only
#   ORTHANC_DEPLOY_ONLY=1 ./infra/gcp/setup-orthanc.sh [IMAGE_URI]
#
# --deploy-only / ORTHANC_DEPLOY_ONLY=1 : met à jour l'image Cloud Run sans
#   recréer bucket/secrets/DB ni reset du mot de passe SQL (chemin CI).
# Sans flag : si service + secrets + DB déjà présents → deploy-only automatique.
#
# Prérequis : Cloud SQL instance, VPC connector, SA petsfollow-run.
# Crée secrets orthanc-* si absents (mots de passe générés) en mode bootstrap.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

IMAGE=""
DEPLOY_ONLY="${ORTHANC_DEPLOY_ONLY:-0}"
for arg in "$@"; do
  case "$arg" in
    --deploy-only) DEPLOY_ONLY=1 ;;
    -*)
      echo "Usage: $0 [IMAGE_URI] [--deploy-only]" >&2
      exit 2
      ;;
    *)
      if [[ -z "$IMAGE" ]]; then
        IMAGE="$arg"
      else
        echo "Usage: $0 [IMAGE_URI] [--deploy-only]" >&2
        exit 2
      fi
      ;;
  esac
done
IMAGE="${IMAGE:-$(ar_image orthanc latest)}"

SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"
CONNECTOR="$(connector_path)"
BUCKET="${GCS_DICOM_BUCKET:-petsfollow-dicom}"
LOCATION="${GCS_MEDIA_LOCATION:-${GCP_RUN_REGION}}"
ORTHANC_DB="${ORTHANC_DB_NAME:-orthanc}"
ORTHANC_DB_USER="${ORTHANC_DB_USER:-orthanc}"
SQL_INSTANCE_SHORT="${SQL_INSTANCE:-premedica-db-staging}"
# Plugin GoogleCloudStorage exige un JSON keyfile (pas d'ADC) — voir GoogleStoragePlugin.cpp.
GCS_SA_SECRET="${GCS_SA_SECRET:-petsfollow-orthanc-gcs-sa}"
GCS_SA_FILE_PATH="${GCS_SA_FILE_PATH:-/var/run/secrets/petsfollow-gcs-sa.json}"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

orthanc_bootstrapped() {
  gcloud run services describe "$ORTHANC_SERVICE" \
    --project="$GCP_PROJECT_ID" --region="$GCP_RUN_REGION" >/dev/null 2>&1 \
  && gcloud secrets describe petsfollow-orthanc-password \
    --project="$GCP_PROJECT_ID" >/dev/null 2>&1 \
  && gcloud secrets describe petsfollow-orthanc-db-password \
    --project="$GCP_PROJECT_ID" >/dev/null 2>&1 \
  && gcloud sql databases describe "$ORTHANC_DB" \
    --instance="$SQL_INSTANCE_SHORT" --project="$GCP_PROJECT_ID" >/dev/null 2>&1
}

if [[ "$DEPLOY_ONLY" != "1" && "$DEPLOY_ONLY" != "true" ]]; then
  if orthanc_bootstrapped; then
    echo "→ Orthanc déjà bootstrappé — mode deploy-only (pas de reset SQL)"
    DEPLOY_ONLY=1
  fi
fi

echo "=== petsFollow Orthanc PACS — ${GCP_PROJECT_ID} / ${GCP_RUN_REGION} ==="
echo "Image: ${IMAGE}"
echo "Service: ${ORTHANC_SERVICE}"
echo "Bucket: gs://${BUCKET}"
echo "Mode: $([[ "$DEPLOY_ONLY" == "1" || "$DEPLOY_ONLY" == "true" ]] && echo deploy-only || echo bootstrap)"

resolve_pg_host() {
  local pg_host private_ip
  pg_host="$(gcloud sql instances describe "$SQL_INSTANCE_SHORT" \
    --project="$GCP_PROJECT_ID" \
    --format='value(ipAddresses[0].ipAddress)' 2>/dev/null || true)"
  private_ip="$(gcloud sql instances describe "$SQL_INSTANCE_SHORT" \
    --project="$GCP_PROJECT_ID" \
    --format=json | python3 -c "
import json,sys
meta=json.load(sys.stdin)
for ip in meta.get('ipAddresses') or []:
  if ip.get('type')=='PRIVATE':
    print(ip.get('ipAddress') or '')
    break
")"
  if [[ -n "$private_ip" ]]; then
    pg_host="$private_ip"
  fi
  if [[ -z "$pg_host" ]]; then
    echo "ERREUR: impossible de résoudre l'IP Cloud SQL" >&2
    exit 1
  fi
  echo "$pg_host"
}

# Crée (une fois) une clé JSON pour ${SA_EMAIL} → Secret Manager, montée en fichier
# sur Cloud Run. Le plugin Orthanc GoogleCloudStorage n'accepte pas l'ADC.
ensure_gcs_sa_secret() {
  if gcloud secrets describe "$GCS_SA_SECRET" \
    --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "  Secret ${GCS_SA_SECRET} existe"
  else
    echo "→ Création clé SA ${SA_EMAIL} → secret ${GCS_SA_SECRET}"
    local keyfile
    keyfile="$(mktemp)"
    if ! gcloud iam service-accounts keys create "$keyfile" \
      --iam-account="$SA_EMAIL" \
      --project="$GCP_PROJECT_ID" \
      --quiet; then
      rm -f "$keyfile"
      echo "ERREUR: iam.serviceAccountKeys.create refusé pour ${SA_EMAIL}" >&2
      echo "  Crée manuellement le secret ${GCS_SA_SECRET} (JSON key) puis relance." >&2
      exit 1
    fi
    gcloud secrets create "$GCS_SA_SECRET" \
      --project="$GCP_PROJECT_ID" \
      --data-file="$keyfile" \
      --replication-policy=automatic \
      --quiet
    rm -f "$keyfile"
    echo "  Secret ${GCS_SA_SECRET} créé"
  fi
  if ! gcloud secrets add-iam-policy-binding "$GCS_SA_SECRET" \
    --project="$GCP_PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/secretmanager.secretAccessor" \
    --quiet >/dev/null 2>&1; then
    echo "  WARN: IAM secretAccessor sur ${GCS_SA_SECRET} ignoré (SA sans setIamPolicy)" >&2
  fi
}

deploy_orthanc_run() {
  local pg_host="$1"
  local env_file
  ensure_gcs_sa_secret
  env_file="$(mktemp)"
  # Orthanc Team generateConfiguration.py : *_PLUGIN_ENABLED symlink les .so
  # depuis plugins-available → plugins/. ORTHANC__* surcharge petsfollow.json.
  # GCS : JSON key monté (${GCS_SA_FILE_PATH}) — pas d'ADC (limitation plugin).
  cat >"$env_file" <<EOF
VERBOSE_ENABLED: "true"
POSTGRESQL_PLUGIN_ENABLED: "true"
GOOGLE_CLOUD_STORAGE_PLUGIN_ENABLED: "true"
DICOM_WEB_PLUGIN_ENABLED: "true"
ORTHANC__NAME: "PetsFollowPACS"
ORTHANC__HTTP_PORT: "8080"
ORTHANC__REMOTE_ACCESS_ALLOWED: "true"
ORTHANC__AUTHENTICATION_ENABLED: "false"
ORTHANC__DICOM_SERVER_ENABLED: "false"
ORTHANC__HTTPS_CA_CERTIFICATES: "/etc/ssl/certs/ca-certificates.crt"
ORTHANC__POSTGRESQL__ENABLE_INDEX: "true"
ORTHANC__POSTGRESQL__ENABLE_STORAGE: "false"
ORTHANC__POSTGRESQL__HOST: "${pg_host}"
ORTHANC__POSTGRESQL__PORT: "5432"
ORTHANC__POSTGRESQL__DATABASE: "${ORTHANC_DB}"
ORTHANC__POSTGRESQL__USERNAME: "${ORTHANC_DB_USER}"
ORTHANC__POSTGRESQL__ENABLE_SSL: "false"
ORTHANC__POSTGRESQL__LOCK: "false"
ORTHANC__POSTGRESQL__INDEX_CONNECTIONS_COUNT: "4"
ORTHANC__POSTGRESQL__MAXIMUM_CONNECTION_RETRIES: "10"
ORTHANC__POSTGRESQL__CONNECTION_RETRY_INTERVAL: "5"
ORTHANC__GOOGLE_CLOUD_STORAGE__BUCKET_NAME: "${BUCKET}"
ORTHANC__GOOGLE_CLOUD_STORAGE__ROOT_PATH: "dicom/"
ORTHANC__GOOGLE_CLOUD_STORAGE__SERVICE_ACCOUNT_FILE: "${GCS_SA_FILE_PATH}"
EOF

  # ingress=all + IAM (no unauthenticated) : l'API Cloud Run appelle Orthanc via
  # URL run.app + ID token. ingress=internal + egress private-ranges-only → 404
  # (trafic run.app sort en public et n'est pas traité comme appel inter-services).
  echo "→ Deploy Cloud Run ${ORTHANC_SERVICE} (min=0 max=10 ingress=all IAM-only)"
  gcloud run deploy "$ORTHANC_SERVICE" \
    --project="$GCP_PROJECT_ID" \
    --image="$IMAGE" \
    --region="$GCP_RUN_REGION" \
    --service-account="$SA_EMAIL" \
    --memory=1Gi --cpu=1 \
    --min-instances=0 --max-instances=10 \
    --timeout=300 \
    --concurrency=20 \
    --cpu-boost \
    --port=8080 \
    --ingress=all \
    --no-allow-unauthenticated \
    --vpc-connector="$CONNECTOR" \
    --vpc-egress=private-ranges-only \
    --set-cloudsql-instances="$CLOUDSQL_INSTANCE" \
    --env-vars-file="$env_file" \
    --set-secrets="ORTHANC__POSTGRESQL__PASSWORD=petsfollow-orthanc-db-password:latest,ORTHANC_PASSWORD=petsfollow-orthanc-password:latest,${GCS_SA_FILE_PATH}=${GCS_SA_SECRET}:latest" \
    --quiet
  rm -f "$env_file"

  # Best-effort : Cloud Build compute SA often lacks run.services.setIamPolicy.
  # petsfollow-run invoker is granted once at bootstrap (or manually).
  if ! gcloud run services add-iam-policy-binding "$ORTHANC_SERVICE" \
    --project="$GCP_PROJECT_ID" \
    --region="$GCP_RUN_REGION" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/run.invoker" \
    --quiet >/dev/null 2>&1; then
    echo "  WARN: IAM run.invoker sur ${ORTHANC_SERVICE} ignoré (SA sans setIamPolicy)" >&2
  fi
}

if [[ "$DEPLOY_ONLY" == "1" || "$DEPLOY_ONLY" == "true" ]]; then
  # IAM bucket best-effort (no create / no password reset).
  gcloud storage buckets add-iam-policy-binding "gs://${BUCKET}" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/storage.objectAdmin" \
    --quiet >/dev/null 2>&1 || true
  PG_HOST="$(resolve_pg_host)"
  echo "  PG_HOST=${PG_HOST}"
  deploy_orthanc_run "$PG_HOST"
else
  # --- Bucket DICOM (privé, versioning) ---
  if gcloud storage buckets describe "gs://${BUCKET}" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "  Bucket gs://${BUCKET} existe déjà"
  else
    gcloud storage buckets create "gs://${BUCKET}" \
      --project="$GCP_PROJECT_ID" \
      --location="$LOCATION" \
      --uniform-bucket-level-access \
      --public-access-prevention \
      --quiet
    echo "  Bucket gs://${BUCKET} créé"
  fi

  gcloud storage buckets update "gs://${BUCKET}" \
    --project="$GCP_PROJECT_ID" \
    --versioning \
    --quiet >/dev/null || true

  gcloud storage buckets add-iam-policy-binding "gs://${BUCKET}" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/storage.objectAdmin" \
    --quiet >/dev/null

  echo "→ Bucket DICOM OK (versioning on, privé)"

  ensure_secret() {
    local name="$1"
    local value="$2"
    if gcloud secrets describe "$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
      echo "  Secret ${name} existe"
    else
      printf '%s' "$value" | gcloud secrets create "$name" \
        --project="$GCP_PROJECT_ID" \
        --data-file=- \
        --replication-policy=automatic \
        --quiet
      echo "  Secret ${name} créé"
    fi
    # Best-effort : Cloud Build SA (compute) peut créer le secret sans setIamPolicy.
    # petsfollow-run a déjà secretAccessor au niveau projet (setup-gcp.sh).
    if ! gcloud secrets add-iam-policy-binding "$name" \
      --project="$GCP_PROJECT_ID" \
      --member="serviceAccount:${SA_EMAIL}" \
      --role="roles/secretmanager.secretAccessor" \
      --quiet >/dev/null 2>&1; then
      echo "  WARN: IAM secretAccessor sur ${name} ignoré (SA Cloud Build sans setIamPolicy)" >&2
    fi
  }

  ORTHANC_PASS="$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)"
  ORTHANC_DB_PASS="$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)"
  ensure_secret "petsfollow-orthanc-password" "$ORTHANC_PASS"
  ensure_secret "petsfollow-orthanc-db-password" "$ORTHANC_DB_PASS"

  ORTHANC_DB_PASS="$(gcloud secrets versions access latest \
    --secret=petsfollow-orthanc-db-password --project="$GCP_PROJECT_ID")"

  echo "→ Secrets Orthanc OK"

  echo "→ Ensure Cloud SQL database/user ${ORTHANC_DB}"
  if gcloud sql databases describe "$ORTHANC_DB" \
    --instance="$SQL_INSTANCE_SHORT" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "  DB ${ORTHANC_DB} existe"
  else
    gcloud sql databases create "$ORTHANC_DB" \
      --instance="$SQL_INSTANCE_SHORT" \
      --project="$GCP_PROJECT_ID" \
      --quiet
  fi

  if gcloud sql users list --instance="$SQL_INSTANCE_SHORT" --project="$GCP_PROJECT_ID" \
    --format='value(name)' | grep -qx "$ORTHANC_DB_USER"; then
    echo "  User ${ORTHANC_DB_USER} existe — maj mot de passe (bootstrap)"
    gcloud sql users set-password "$ORTHANC_DB_USER" \
      --instance="$SQL_INSTANCE_SHORT" \
      --project="$GCP_PROJECT_ID" \
      --password="$ORTHANC_DB_PASS" \
      --quiet
  else
    gcloud sql users create "$ORTHANC_DB_USER" \
      --instance="$SQL_INSTANCE_SHORT" \
      --project="$GCP_PROJECT_ID" \
      --password="$ORTHANC_DB_PASS" \
      --quiet
  fi

  PG_HOST="$(resolve_pg_host)"
  echo "  PG_HOST=${PG_HOST}"
  deploy_orthanc_run "$PG_HOST"
fi

ORTHANC_URL="$(gcloud run services describe "$ORTHANC_SERVICE" \
  --project="$GCP_PROJECT_ID" --region="$GCP_RUN_REGION" \
  --format='value(status.url)')"

echo ""
echo "Orthanc déployé : ${ORTHANC_URL}"
echo "Configurer l'API Cloud Run :"
echo "  PACS_ENABLED=true"
echo "  PACS_ORTHANC_URL=${ORTHANC_URL}"
echo "  PACS_ORTHANC_USER=petsfollow"
echo "  PACS_ORTHANC_PASSWORD=<secret petsfollow-orthanc-password>"
echo "  PACS_ORTHANC_USE_ID_TOKEN=true"
echo ""
echo "Backups : Cloud SQL automated backups + GCS versioning sur gs://${BUCKET}"
echo "Plugins : POSTGRESQL + GOOGLE_CLOUD_STORAGE + DICOM_WEB (entrypoint Orthanc Team)"
echo "GCS key : secret ${GCS_SA_SECRET} → ${GCS_SA_FILE_PATH}"
echo "Pooling : IndexConnectionsCount=4 — surveiller max_connections Cloud SQL"
echo "Vérif logs : Registering PostgreSQL/GoogleCloud (pas SQLite+/tmp FilesystemStorage)"
