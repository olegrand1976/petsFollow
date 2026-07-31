#!/usr/bin/env bash
# Enregistre l'app Flutter petsFollow sur Firebase (projet GCP partagé).
# Auth : PostgreSQL via API Go — Firebase Auth volontairement NON utilisé.
# Flavors Android : staging (.staging) + prod (package store).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FLUTTER_DIR="${ROOT}/flutter"
PROJECT_ID="${GCP_PROJECT_ID:-premedica-prod-2025}"
ANDROID_PACKAGE_PROD="be.llitsc.petsfollow_mobile"
ANDROID_PACKAGE_STAGING="be.llitsc.petsfollow_mobile.staging"
IOS_BUNDLE_ID="be.llitsc.petsfollowMobile"

echo "→ Projet Firebase/GCP : ${PROJECT_ID}"
echo "→ Android prod     : ${ANDROID_PACKAGE_PROD}"
echo "→ Android staging  : ${ANDROID_PACKAGE_STAGING}"
echo "→ iOS              : ${IOS_BUNDLE_ID}"

cd "${FLUTTER_DIR}"

if [[ ! -d android/app ]]; then
  echo "→ Génération plateformes android/ios"
  flutter create . --org be.llitsc --project-name petsfollow_mobile --platforms=android,ios
fi

echo "→ Création apps Firebase (idempotent — échoue si déjà existantes)"
firebase apps:create android "petsFollow pets" \
  --package-name="${ANDROID_PACKAGE_PROD}" \
  --project="${PROJECT_ID}" 2>/dev/null || true

firebase apps:create android "petsFollow pets (staging)" \
  --package-name="${ANDROID_PACKAGE_STAGING}" \
  --project="${PROJECT_ID}" 2>/dev/null || true

firebase apps:create ios "petsFollow pets (iOS)" \
  --bundle-id="${IOS_BUNDLE_ID}" \
  --project="${PROJECT_ID}" 2>/dev/null || true

resolve_android_app_id() {
  local pkg="$1"
  firebase apps:list ANDROID --project="${PROJECT_ID}" --json 2>/dev/null \
    | python3 -c "
import json, sys
data = json.load(sys.stdin)
apps = data.get('result', data) if isinstance(data, dict) else data
if isinstance(apps, dict):
    apps = apps.get('apps', apps.get('result', []))
for a in apps:
    if a.get('packageName') == '''${pkg}''' or a.get('package_name') == '''${pkg}''':
        print(a.get('appId') or a.get('name', '').split('/')[-1])
        break
" 2>/dev/null || true
}

# Fallback known IDs if JSON parse fails
PROD_APP_ID="${FIREBASE_ANDROID_APP_ID_PROD:-1:237481297060:android:cfda5c59a08bfd6dc9d231}"
STAGING_APP_ID="${FIREBASE_ANDROID_APP_ID_STAGING:-1:237481297060:android:a15f2666e4707cf1c9d231}"
RESOLVED_PROD="$(resolve_android_app_id "${ANDROID_PACKAGE_PROD}" || true)"
RESOLVED_STAGING="$(resolve_android_app_id "${ANDROID_PACKAGE_STAGING}" || true)"
[[ -n "${RESOLVED_PROD}" ]] && PROD_APP_ID="${RESOLVED_PROD}"
[[ -n "${RESOLVED_STAGING}" ]] && STAGING_APP_ID="${RESOLVED_STAGING}"

write_sdkconfig() {
  local app_id="$1"
  local out="$2"
  mkdir -p "$(dirname "${out}")"
  local tmp
  tmp="$(mktemp)"
  firebase apps:sdkconfig ANDROID "${app_id}" --project="${PROJECT_ID}" >"${tmp}"
  python3 - "${tmp}" "${out}" <<'PY'
import json, sys
from pathlib import Path
raw = Path(sys.argv[1]).read_text()
start, end = raw.find('{'), raw.rfind('}')
data = json.loads(raw[start:end+1])
Path(sys.argv[2]).write_text(json.dumps(data, indent=2) + '\n')
PY
  rm -f "${tmp}"
  echo "  ✓ ${out} (app ${app_id})"
}

echo "→ google-services.json par flavor"
write_sdkconfig "${PROD_APP_ID}" "${FLUTTER_DIR}/android/app/src/prod/google-services.json"
write_sdkconfig "${STAGING_APP_ID}" "${FLUTTER_DIR}/android/app/src/staging/google-services.json"

echo "→ FlutterFire configure (prod package — firebase_options.dart)"
flutterfire configure \
  --project="${PROJECT_ID}" \
  --platforms=android,ios \
  --android-package-name="${ANDROID_PACKAGE_PROD}" \
  --ios-bundle-id="${IOS_BUNDLE_ID}" \
  --yes \
  --out=lib/firebase_options.dart

echo "→ Dépendance firebase_core"
flutter pub add firebase_core 2>/dev/null || true

echo ""
echo "✓ Firebase Flutter prêt (flavors staging + prod)."
echo "  Staging App Distribution ID : ${STAGING_APP_ID}"
echo "  Prod Play package           : ${ANDROID_PACKAGE_PROD}"
echo "  Auth : POST /api/v1/auth/login → JWT (PostgreSQL)"
echo "  Firebase : FCM / infra mobile uniquement — pas Firebase Auth"
echo ""
echo "Prochaines étapes manuelles (console Firebase) :"
echo "  1. Vérifier que Authentication n'a AUCUN provider activé pour ces apps"
echo "  2. SHA staging : make firebase-google-signin-android SHA1=… FLAVOR=staging"
echo "  3. Activer Cloud Messaging quand FCM sera branché côté API Go"
