#!/usr/bin/env bash
# Provisionne le secret AFMPS VAMREG readonly (FAMHP-SEC-KEY) + IAM Cloud Run.
#
# Usage:
#   ./infra/gcp/setup-vamreg-afmps-secret.sh /chemin/vers/fichier-cle
#
# Le fichier doit contenir UNIQUEMENT la clé API (une ligne, format FAMHP
# uuid_…=) — pas de KEY=value (le « = » final fait partie de la clé).
#
# Puis monter sur le service API (si pas déjà via prochain gcp-deploy) :
#   ./infra/gcp/setup-vamreg-afmps-secret.sh --attach-run
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

SECRET_NAME="petsfollow-vamreg-afmps-api-key"
DEFAULT_BASE_URL="https://app.fagg-afmps.be/vamreg/api"
SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"

ATTACH_RUN=false
KEY_FILE=""
for arg in "$@"; do
  case "$arg" in
    --attach-run) ATTACH_RUN=true ;;
    -*)
      echo "Unknown flag: $arg" >&2
      exit 2
      ;;
    *)
      KEY_FILE="$arg"
      ;;
  esac
done

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

ensure_secret_from_file() {
  local file="$1"
  if [[ ! -f "$file" ]]; then
    echo "Fichier introuvable: $file" >&2
    exit 1
  fi
  local tmp
  tmp="$(mktemp)"
  # Trim trailing newlines only — preserve trailing '=' of FAMHP keys.
  python3 - "$file" "$tmp" <<'PY'
import pathlib, sys
src, dst = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
raw = src.read_bytes().strip()
if not raw:
    raise SystemExit("empty key file")
# Reject accidental KEY=value wrappers (except bare key ending with =).
text = raw.decode("utf-8")
if "\n" in text:
    raise SystemExit("key file must be a single line")
dst.write_bytes(raw)
print(f"key_bytes={len(raw)}")
PY
  if gcloud secrets describe "$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "→ ADD version ${SECRET_NAME}"
    gcloud secrets versions add "$SECRET_NAME" --data-file="$tmp" --project="$GCP_PROJECT_ID" --quiet
  else
    echo "→ CREATE ${SECRET_NAME}"
    gcloud secrets create "$SECRET_NAME" \
      --replication-policy=automatic \
      --data-file="$tmp" \
      --project="$GCP_PROJECT_ID" --quiet
  fi
  rm -f "$tmp"
  echo "→ IAM secretAccessor → ${SA_EMAIL}"
  gcloud secrets add-iam-policy-binding "$SECRET_NAME" \
    --project="$GCP_PROJECT_ID" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/secretmanager.secretAccessor" \
    --quiet >/dev/null
}

attach_run() {
  echo "→ Cloud Run ${API_SERVICE} : secrets + base URL AFMPS + DRY_RUN=true"
  gcloud run services update "$API_SERVICE" \
    --project="$GCP_PROJECT_ID" \
    --region="$GCP_RUN_REGION" \
    --update-secrets="VAMREG_AFMPS_API_KEY=${SECRET_NAME}:latest" \
    --update-env-vars="VAMREG_AFMPS_BASE_URL=${DEFAULT_BASE_URL},VAMREG_DRY_RUN=true" \
    --quiet
  echo "✓ Attach OK (déclarations dry-run ; listes AFMPS live GET)"
}

if [[ -n "$KEY_FILE" ]]; then
  ensure_secret_from_file "$KEY_FILE"
fi

if [[ "$ATTACH_RUN" == true ]]; then
  if ! gcloud secrets versions access latest --secret="$SECRET_NAME" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "Secret ${SECRET_NAME} absent — fournis un fichier clé avant --attach-run" >&2
    exit 1
  fi
  attach_run
fi

if [[ -z "$KEY_FILE" && "$ATTACH_RUN" != true ]]; then
  echo "Usage: $0 /chemin/cle [--attach-run]" >&2
  echo "       $0 --attach-run   # monte un secret déjà créé" >&2
  exit 2
fi

echo "Done."
