#!/usr/bin/env bash
# Build Android App Bundle (AAB) for Google Play Console upload — flavor prod.
# Requires flutter/android/key.properties + upload keystore (see key.properties.example).
# Until a dedicated prod API exists, pass API_BASE explicitly (HTTPS).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FLUTTER_DIR="${ROOT}/flutter"
ANDROID_DIR="${FLUTTER_DIR}/android"
# No silent default to staging Cloud Run for prod-flavored store builds.
if [[ -z "${API_BASE:-}" ]]; then
  echo "ERROR: set API_BASE=https://… explicitly for prod AAB builds." >&2
  echo "  (Staging testeurs : make firebase-android-dist — flavor staging.)" >&2
  exit 1
fi
GOOGLE_SERVER_CLIENT_ID="${GOOGLE_SERVER_CLIENT_ID:-237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com}"

if [[ "${API_BASE}" != https://* ]]; then
  echo "ERROR: API_BASE must be https:// for Play release (got: ${API_BASE})" >&2
  exit 1
fi

if [[ ! -f "${ANDROID_DIR}/key.properties" ]]; then
  echo "ERROR: missing ${ANDROID_DIR}/key.properties" >&2
  echo "Copy key.properties.example → key.properties and point storeFile to your upload-keystore.jks" >&2
  exit 1
fi

STORE_FILE_PROP="$(grep -E '^storeFile=' "${ANDROID_DIR}/key.properties" | head -1 | cut -d= -f2-)"
if [[ -z "${STORE_FILE_PROP}" ]]; then
  echo "ERROR: key.properties missing storeFile=" >&2
  exit 1
fi
# storeFile is relative to flutter/android/
STORE_FILE_PATH="${ANDROID_DIR}/${STORE_FILE_PROP}"
if [[ ! -f "${STORE_FILE_PATH}" ]]; then
  echo "ERROR: keystore not found: ${STORE_FILE_PATH}" >&2
  exit 1
fi

cd "${FLUTTER_DIR}"

PUBSPEC="${FLUTTER_DIR}/pubspec.yaml"
CURRENT="$(grep -E '^version:' "${PUBSPEC}" | head -1 | awk '{print $2}')"
echo "→ Building Play App Bundle flavor=prod APP_ENV=prod (version ${CURRENT})"
echo "→ API_BASE=${API_BASE}"
echo "→ Package be.llitsc.petsfollow_mobile (pas .staging)"

flutter pub get
flutter build appbundle --release --flavor prod \
  --dart-define="FLAVOR=prod" \
  --dart-define="APP_ENV=prod" \
  --dart-define="API_BASE=${API_BASE}" \
  --dart-define="GOOGLE_SERVER_CLIENT_ID=${GOOGLE_SERVER_CLIENT_ID}"

AAB_PATH="${FLUTTER_DIR}/build/app/outputs/bundle/prodRelease/app-prod-release.aab"
if [[ ! -f "${AAB_PATH}" ]]; then
  AAB_PATH="${FLUTTER_DIR}/build/app/outputs/bundle/release/app-release.aab"
fi
test -f "${AAB_PATH}"

echo ""
echo "✓ AAB prêt : ${AAB_PATH}"
echo "  Upload → Play Console → Testing (internal/closed) puis Production."
echo "  Privacy policy URL : https://petsfollow.ll-it-sc.be/legal/privacy"
echo "  Checklist : documentation/26-PLAY-STORE.md"
echo "  Canal testeurs staging : make firebase-android-dist"
