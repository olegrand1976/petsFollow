#!/usr/bin/env bash
# Enregistre des SHA Android sur Firebase petsFollow et régénère google-services.json.
# Usage :
#   ./infra/firebase/setup-google-signin-android.sh SHA1 [SHA256 ...]
#   FLAVOR=staging|prod ./infra/firebase/setup-google-signin-android.sh SHA1
# Exemple (debug) :
#   CERT=$(mktemp)
#   keytool -exportcert -alias androiddebugkey -keystore ~/.android/debug.keystore \
#     -storepass android -rfc >"$CERT"
#   SHA1=$(openssl x509 -fingerprint -sha1 -noout -in "$CERT" | cut -d= -f2)
#   rm -f "$CERT"
#   FLAVOR=staging ./infra/firebase/setup-google-signin-android.sh "$SHA1"
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROJECT_ID="${GCP_PROJECT_ID:-premedica-prod-2025}"
FLAVOR="${FLAVOR:-staging}"

case "${FLAVOR}" in
  staging)
    APP_ID="${FIREBASE_ANDROID_APP_ID:-1:237481297060:android:a15f2666e4707cf1c9d231}"
    OUT="${ROOT}/flutter/android/app/src/staging/google-services.json"
    PKG="be.llitsc.petsfollow_mobile.staging"
    ;;
  prod)
    APP_ID="${FIREBASE_ANDROID_APP_ID_PROD:-1:237481297060:android:cfda5c59a08bfd6dc9d231}"
    OUT="${ROOT}/flutter/android/app/src/prod/google-services.json"
    PKG="be.llitsc.petsfollow_mobile"
    ;;
  *)
    echo "FLAVOR must be staging or prod (got: ${FLAVOR})" >&2
    exit 1
    ;;
esac

if [[ $# -lt 1 ]]; then
  echo "Usage: FLAVOR=staging|prod $0 SHA_HASH [SHA_HASH ...]" >&2
  exit 1
fi

echo "→ Projet ${PROJECT_ID} / flavor ${FLAVOR} / app ${APP_ID}"
for sha in "$@"; do
  echo "→ Add SHA ${sha}"
  firebase apps:android:sha:create "$APP_ID" "$sha" --project="$PROJECT_ID" 2>/dev/null || true
done

echo "→ Liste SHA"
firebase apps:android:sha:list "$APP_ID" --project="$PROJECT_ID"

echo "→ Régénération google-services.json → ${OUT}"
mkdir -p "$(dirname "${OUT}")"
tmp="$(mktemp)"
firebase apps:sdkconfig ANDROID "$APP_ID" --project="$PROJECT_ID" >"$tmp"
python3 - "$tmp" "$OUT" "$PKG" <<'PY'
import json, sys
from pathlib import Path
raw = Path(sys.argv[1]).read_text()
start, end = raw.find('{'), raw.rfind('}')
data = json.loads(raw[start:end+1])
Path(sys.argv[2]).write_text(json.dumps(data, indent=2) + '\n')
pkg = sys.argv[3]
for c in data.get('client', []):
    if c.get('client_info', {}).get('android_client_info', {}).get('package_name') == pkg:
        types = [o.get('client_type') for o in c.get('oauth_client', [])]
        print('oauth_client types for', pkg, ':', types)
        if 1 not in types:
            raise SystemExit('✗ Pas de client OAuth Android (type 1) — SHA non propagé ?')
        print('✓ Client OAuth Android présent')
        break
else:
    raise SystemExit(f'✗ Package {pkg} absent du google-services.json')
PY
rm -f "$tmp"
echo "✓ ${OUT}"
