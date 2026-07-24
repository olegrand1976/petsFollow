#!/usr/bin/env bash
# Postdeploy petsFollow : jobs + seed optionnel + smoke.
# Usage: ./infra/gcp/postdeploy.sh [--seed]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

RUN_SEED=false
for arg in "$@"; do
  case "$arg" in
    --seed|--seed-reset) RUN_SEED=true ;;
    --skip-seed) RUN_SEED=false ;;
  esac
done

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

run_job() {
  local name="$1"
  echo "→ Exécution ${name}"
  gcloud run jobs execute "$name" \
    --region="$GCP_RUN_REGION" --project="$GCP_PROJECT_ID" \
    --wait --quiet
}

echo "=== petsFollow postdeploy — ${GCP_PROJECT_ID} ==="

bash "${SCRIPT_DIR}/setup-jobs.sh"
bash "${SCRIPT_DIR}/grant-app-privileges.sh"

if $RUN_SEED; then
  run_job "petsfollow-seed"
fi

API_URL="$(api_run_url)"
FE_URL="$(frontend_run_url)"
export PETSFOLLOW_API_URL="${PUBLIC_API_URL}"
if ! curl -sf --max-time 5 "${PUBLIC_API_URL}/health" >/dev/null 2>&1; then
  echo "→ Domaine custom pas encore prêt — smoke via Cloud Run URL"
  export PETSFOLLOW_API_URL="$API_URL"
fi

bash "$(cd "${SCRIPT_DIR}/../.." && pwd)/scripts/smoke-test.sh"

cat <<EOF

Postdeploy terminé.
  API Run   : ${API_URL:-non déployée}
  Frontend  : ${FE_URL:-non déployé}
  Custom    : ${PUBLIC_SITE_URL} / ${PUBLIC_API_URL}

EOF
