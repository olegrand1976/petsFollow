#!/usr/bin/env bash
# Preflight deploy petsFollow — prérequis GCP + baseline santé avant Cloud Build.
#
# Usage:
#   ./infra/gcp/preflight-deploy.sh --env staging
#   ./infra/gcp/preflight-deploy.sh --env prod
#   SKIP_BASELINE=1 ./infra/gcp/preflight-deploy.sh --env staging
#
# Gates CI (staging) : quality (ci.yml) → preflight → Cloud Build → postdeploy smoke → Playwright
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PF_ENV=""
SKIP_BASELINE="${SKIP_BASELINE:-0}"

usage() {
  echo "Usage: $0 --env staging|prod [--skip-baseline]" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env)
      PF_ENV="${2:-}"
      shift 2
      ;;
    --skip-baseline)
      SKIP_BASELINE=1
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Argument inconnu: $1" >&2
      usage
      ;;
  esac
done

[[ -n "$PF_ENV" ]] || usage

fail() {
  echo "ERREUR preflight: $*" >&2
  exit 1
}

check_secret() {
  local secret_name="$1"
  if ! gcloud secrets versions access latest \
    --secret="$secret_name" \
    --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    fail "Secret inaccessible ou sans version active: ${secret_name}"
  fi
  echo "OK secret ${secret_name}"
}

check_cloud_run_ready() {
  local service="$1"
  # Premier bootstrap : service absent → WARN (deploy autorisé). Service présent
  # mais non Ready → fail (évite d'empiler sur une révision déjà cassée).
  if ! gcloud run services describe "$service" \
    --region="$GCP_RUN_REGION" \
    --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "WARN Cloud Run ${service} introuvable — premier deploy / bootstrap autorisé"
    return 0
  fi
  local ready
  ready="$(gcloud run services describe "$service" \
    --region="$GCP_RUN_REGION" \
    --project="$GCP_PROJECT_ID" \
    --format=json 2>/dev/null | python3 -c "
import json, sys
data = json.load(sys.stdin)
conditions = data.get('status', {}).get('conditions', [])
ready_status = next((c.get('status') for c in conditions if c.get('type') == 'Ready'), 'Unknown')
print(ready_status)
")"
  if [[ "$ready" != "True" ]]; then
    fail "Cloud Run ${service} non Ready (status=${ready})"
  fi
  echo "OK Cloud Run ${service} Ready"
}

check_cloudbuild_core() {
  local file="$1"
  [[ -f "$file" ]] || fail "cloudbuild introuvable: ${file}"
  # Les env DATABASE_URL / JWT_SIGNING_KEY ne sont pas en clair dans le YAML :
  # elles sont assemblées à runtime par pf_api_secrets (deploy-run-args.sh).
  for needle in petsfollow-api petsfollow-nuxtjs pf_api_secrets --set-secrets; do
    grep -qF -- "$needle" "$file" || fail "${file} : ${needle} manquant"
  done
  echo "OK cloudbuild core (${file##*/})"
}

check_baseline_http() {
  local label="$1"
  local url="$2"
  local allowed="${3:-200}"
  local code
  code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 30 "$url" || echo "000")"
  if [[ "$allowed" == *"$code"* ]]; then
    echo "OK baseline ${label} (${code})"
    return 0
  fi
  # Env déjà cassé : autoriser le redeploy de recovery (comme Infiswap).
  if [[ "$code" == "502" || "$code" == "503" || "$code" == "504" ]]; then
    echo "WARN baseline ${label} HTTP ${code} — env dégradé, deploy recovery autorisé (${url})"
    return 0
  fi
  fail "baseline ${label} HTTP ${code} (attendu ${allowed}) — ${url}"
}

case "$PF_ENV" in
  staging)
    # shellcheck source=lib/gcp-env.sh
    source "${SCRIPT_DIR}/lib/gcp-env.sh"
    CLOUDBUILD_FILE="${SCRIPT_DIR}/cloudbuild.yaml"
    SQL_NEEDLE="premedica-db-staging"
    SECRETS=(
      petsfollow-database-url
      petsfollow-migrate-database-url
      petsfollow-jwt-signing-key
      petsfollow-redis-url
      petsfollow-bff-proxy-secret
      petsfollow-retention-secret
    )
    ;;
  prod)
    # shellcheck source=lib/gcp-env-prod.sh
    source "${SCRIPT_DIR}/lib/gcp-env-prod.sh"
    CLOUDBUILD_FILE="${SCRIPT_DIR}/cloudbuild-prod.yaml"
    SQL_NEEDLE="petsfollow-db-prod"
    SECRETS=(
      petsfollow-prod-database-url
      petsfollow-prod-migrate-database-url
      petsfollow-prod-jwt-signing-key
      petsfollow-prod-redis-url
      petsfollow-prod-bff-proxy-secret
      petsfollow-prod-retention-secret
    )
    ;;
  *)
    fail "env inconnu: ${PF_ENV} (staging|prod)"
    ;;
esac

API_PUBLIC_URL="${PETSFOLLOW_API_PUBLIC_URL:-${PUBLIC_API_URL}}"
WEB_PUBLIC_URL="${PETSFOLLOW_PUBLIC_SITE_URL:-${PUBLIC_SITE_URL}}"

echo "=== Preflight deploy petsFollow ${PF_ENV} ==="
echo "  Projet : ${GCP_PROJECT_ID}"
echo "  API    : ${API_PUBLIC_URL}"
echo "  Web    : ${WEB_PUBLIC_URL}"
echo "  Run    : ${API_SERVICE} / ${FRONTEND_SERVICE}"
echo ""

echo "--- Secrets Secret Manager ---"
for secret in "${SECRETS[@]}"; do
  check_secret "$secret"
done

echo ""
echo "--- Ressources GCP ---"

if ! gcloud run services describe "$API_SERVICE" \
  --region="$GCP_RUN_REGION" \
  --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "WARN Cloud SQL check skip — ${API_SERVICE} absent (premier deploy)"
else
  sql_state="$(gcloud run services describe "$API_SERVICE" \
    --region="$GCP_RUN_REGION" \
    --project="$GCP_PROJECT_ID" \
    --format=json 2>/dev/null | python3 -c "
import json, sys
data = json.load(sys.stdin)
instances = data.get('spec', {}).get('template', {}).get('metadata', {}).get('annotations', {}).get('run.googleapis.com/cloudsql-instances', '')
print('ok' if '${SQL_NEEDLE}' in instances else 'missing')
" || echo "missing")"
  [[ "$sql_state" == "ok" ]] || fail "Cloud SQL ${SQL_NEEDLE} non rattaché à ${API_SERVICE}"
  echo "OK Cloud SQL ${SQL_NEEDLE} (via ${API_SERVICE})"

  if ! gcloud compute networks vpc-access connectors describe "$VPC_CONNECTOR" \
    --region="$GCP_RUN_REGION" \
    --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    vpc_ok="$(gcloud run services describe "$API_SERVICE" \
      --region="$GCP_RUN_REGION" \
      --project="$GCP_PROJECT_ID" \
      --format=json 2>/dev/null | python3 -c "
import json, sys
data = json.load(sys.stdin)
connector = data.get('spec', {}).get('template', {}).get('metadata', {}).get('annotations', {}).get('run.googleapis.com/vpc-access-connector', '')
print('ok' if '${VPC_CONNECTOR}' in connector else 'missing')
" || echo "missing")"
    [[ "$vpc_ok" == "ok" ]] || fail "VPC connector ${VPC_CONNECTOR} introuvable"
  fi
  echo "OK VPC connector ${VPC_CONNECTOR}"
fi

if ! gcloud artifacts repositories describe "$AR_REPO" \
  --location="$GCP_AR_REGION" \
  --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  fail "Artifact Registry introuvable: ${AR_REPO} (${GCP_AR_REGION})"
fi
echo "OK Artifact Registry ${AR_REPO}"

check_cloud_run_ready "$API_SERVICE"
check_cloud_run_ready "$FRONTEND_SERVICE"

echo ""
echo "--- Config cloudbuild ---"
check_cloudbuild_core "$CLOUDBUILD_FILE"

if [[ "$SKIP_BASELINE" == "1" ]]; then
  echo ""
  echo "SKIP baseline (--skip-baseline / SKIP_BASELINE=1)"
else
  echo ""
  echo "--- Baseline santé (env actuel) ---"
  # /health et /ready sont hors préfixe /api/v1 sur l'API Go.
  check_baseline_http "API /health" "${API_PUBLIC_URL}/health" "200"
  check_baseline_http "API /ready" "${API_PUBLIC_URL}/ready" "200"
  check_baseline_http "Web homepage" "${WEB_PUBLIC_URL}/" "200 304"
fi

echo ""
echo "Preflight ${PF_ENV} réussi."
