#!/usr/bin/env bash
# Postdeploy petsFollow : jobs + seed optionnel + smoke.
# Usage: ./infra/gcp/postdeploy.sh [--seed] [--seed-mass]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

RUN_SEED=false
RUN_SEED_MASS=false
for arg in "$@"; do
  case "$arg" in
    --seed|--seed-reset) RUN_SEED=true ;;
    --seed-mass) RUN_SEED_MASS=true ;;
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
if $RUN_SEED_MASS; then
  run_job "petsfollow-seed-mass"
fi

API_URL="$(api_run_url)"
FE_URL="$(frontend_run_url)"
export PETSFOLLOW_API_URL="${PUBLIC_API_URL}"
if ! curl -sf --max-time 5 "${PUBLIC_API_URL}/health" >/dev/null 2>&1; then
  echo "→ Domaine custom pas encore prêt — smoke via Cloud Run URL"
  export PETSFOLLOW_API_URL="$API_URL"
fi

# Purge des artefacts smoke même si le smoke échoue (PF_SKIP_QUALITY_CLEANUP=1
# pour la déléguer, ex. job cleanup-quality du workflow staging).
run_quality_cleanup() {
  if [[ "${PF_SKIP_QUALITY_CLEANUP:-}" == "1" ]]; then
    return 0
  fi
  echo "→ Purge artefacts smoke/e2e (staging DB)"
  bash "${SCRIPT_DIR}/cleanup-staging-quality.sh" \
    || echo "WARN: purge quality échouée — relancer make gcp-staging-quality-cleanup" >&2
}
trap run_quality_cleanup EXIT

bash "$(cd "${SCRIPT_DIR}/../.." && pwd)/scripts/smoke-test.sh"

cat <<EOF

Postdeploy terminé.
  API Run   : ${API_URL:-non déployée}
  Frontend  : ${FE_URL:-non déployé}
  Custom    : ${PUBLIC_SITE_URL} / ${PUBLIC_API_URL}

EOF
