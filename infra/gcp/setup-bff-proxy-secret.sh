#!/usr/bin/env bash
# Provisionne BFF_PROXY_SECRET (X-PF-Client-IP BFF→API) + montage Cloud Run API + Nuxt.
#
# Staging (défaut) : petsfollow-bff-proxy-secret
# Prod             : PETSFOLLOW_GCP_ENV=prod → petsfollow-prod-bff-proxy-secret
#
# Usage:
#   ./infra/gcp/setup-bff-proxy-secret.sh              # crée si absent (valeur aléatoire)
#   BFF_PROXY_SECRET=… ./infra/gcp/setup-bff-proxy-secret.sh   # crée / ajoute une version
#   ./infra/gcp/setup-bff-proxy-secret.sh --attach-run  # monte sur API + Nuxt
#   ./infra/gcp/setup-bff-proxy-secret.sh --attach-run  # après create (idempotent)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/scheduler-env.sh
source "${SCRIPT_DIR}/lib/scheduler-env.sh"

SECRET_NAME="${PF_SCHED_PREFIX}-bff-proxy-secret"
SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"

ATTACH_RUN=false
for arg in "$@"; do
  case "$arg" in
    --attach-run) ATTACH_RUN=true ;;
    -h|--help)
      echo "Usage: $0 [--attach-run]"
      echo "  BFF_PROXY_SECRET=…  optionnel — sinon génère openssl rand -hex 32 si secret absent"
      echo "  PETSFOLLOW_GCP_ENV=prod  pour le préfixe petsfollow-prod-*"
      exit 0
      ;;
    -*)
      echo "Unknown flag: $arg" >&2
      exit 2
      ;;
    *)
      echo "Unexpected arg: $arg" >&2
      exit 2
      ;;
  esac
done

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

ensure_secret() {
  local value="${BFF_PROXY_SECRET:-}"
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    if [[ -n "$value" ]]; then
      echo "→ ADD version ${SECRET_NAME}"
      echo -n "$value" | gcloud secrets versions add "$SECRET_NAME" \
        --project="$GCP_PROJECT_ID" --data-file=- --quiet
    else
      echo "→ Secret ${SECRET_NAME} déjà présent (pas de nouvelle version — passe BFF_PROXY_SECRET=… pour en ajouter)"
    fi
  else
    if [[ -z "$value" ]]; then
      value="$(openssl rand -hex 32)"
      echo "→ Génération aléatoire (32 octets hex) pour ${SECRET_NAME}"
    fi
    echo "→ CREATE ${SECRET_NAME}"
    echo -n "$value" | gcloud secrets create "$SECRET_NAME" \
      --project="$GCP_PROJECT_ID" \
      --replication-policy=automatic \
      --data-file=- --quiet
  fi

  echo "→ IAM secretAccessor → ${SA_EMAIL}"
  gcloud secrets add-iam-policy-binding "$SECRET_NAME" \
    --project="$GCP_PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/secretmanager.secretAccessor" \
    --quiet >/dev/null
}

attach_run() {
  if ! gcloud secrets versions access latest \
    --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "Secret ${SECRET_NAME} absent — lance d'abord sans --attach-run" >&2
    exit 1
  fi

  echo "→ Cloud Run ${API_SERVICE} : BFF_PROXY_SECRET=${SECRET_NAME}:latest"
  gcloud run services update "$API_SERVICE" \
    --project="$GCP_PROJECT_ID" \
    --region="$GCP_RUN_REGION" \
    --update-secrets="BFF_PROXY_SECRET=${SECRET_NAME}:latest" \
    --quiet

  echo "→ Cloud Run ${FRONTEND_SERVICE} : BFF_PROXY_SECRET=${SECRET_NAME}:latest"
  gcloud run services update "$FRONTEND_SERVICE" \
    --project="$GCP_PROJECT_ID" \
    --region="$GCP_RUN_REGION" \
    --update-secrets="BFF_PROXY_SECRET=${SECRET_NAME}:latest" \
    --quiet

  echo "✓ Attach OK — rate limit Pro utilise X-PF-Client-IP (même secret API + Nuxt)"
}

ensure_secret

if [[ "$ATTACH_RUN" == true ]]; then
  attach_run
else
  echo "→ Pour monter sur Cloud Run maintenant : $0 --attach-run"
  echo "  (sinon le prochain gcp-deploy le fera via pf_api_secrets / pf_nuxt_secrets)"
fi

echo "Done (${SECRET_NAME})."
