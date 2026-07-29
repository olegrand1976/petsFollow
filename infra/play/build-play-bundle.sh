#!/usr/bin/env bash
# Build Android App Bundle (AAB) for Google Play Console upload — flavor prod.
# Requires flutter/android/key.properties + upload keystore (see key.properties.example).
#
# Prefer Make targets:
#   make play-android-bundle-internal  # staging API (ALLOW_STAGING_API=1)
#   make play-android-bundle-prod      # api.petsfollow.app (après deploy main/GCP prod)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=api-bases.sh
source "${ROOT}/infra/play/api-bases.sh"

FLUTTER_DIR="${ROOT}/flutter"
ANDROID_DIR="${FLUTTER_DIR}/android"

# No silent default — pick staging (Internal) or prod explicitly via Make / env.
if [[ -z "${API_BASE:-}" ]]; then
  echo "ERROR: set API_BASE=https://… explicitly for prod AAB builds." >&2
  echo "  Internal testing (staging API) :" >&2
  echo "    make play-android-bundle-internal" >&2
  echo "  Production (api.petsfollow.app, après GCP prod) :" >&2
  echo "    make play-android-bundle-prod" >&2
  echo "  (APK testeurs Firebase : make firebase-android-dist — flavor staging.)" >&2
  exit 1
fi
GOOGLE_SERVER_CLIENT_ID="${GOOGLE_SERVER_CLIENT_ID:-237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com}"

if [[ "${API_BASE}" != https://* ]]; then
  echo "ERROR: API_BASE must be https:// for Play release (got: ${API_BASE})" >&2
  exit 1
fi

# Warn when pointing the store package at the known staging API (seedable).
API_BASE_STRIPPED="${API_BASE%/}"
STAGING_STRIPPED="${PETSFOLLOW_API_STAGING%/}"
PROD_STRIPPED="${PETSFOLLOW_API_PROD%/}"
case "${API_BASE_STRIPPED}" in
  "${STAGING_STRIPPED}"|https://petsfollow-api-a7ako2njea-od.a.run.app)
    echo "WARNING: API_BASE points at the staging API (${API_BASE_STRIPPED})." >&2
    echo "  Suitable for Play Internal testing only — staging is seedable / resettable." >&2
    echo "  For Production store track: make play-android-bundle-prod (${PROD_STRIPPED})." >&2
    if [[ "${ALLOW_STAGING_API:-}" != "1" ]]; then
      echo "  Re-run with ALLOW_STAGING_API=1 (or make play-android-bundle-internal)." >&2
      exit 1
    fi
    ;;
  "${PROD_STRIPPED}")
    echo "→ API_BASE = production (${PROD_STRIPPED})"
    ;;
  *)
    echo "→ API_BASE = custom HTTPS (${API_BASE_STRIPPED})"
    ;;
esac

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

if [[ "${SKIP_TESTS:-}" != "1" ]]; then
  echo "→ flutter test (SKIP_TESTS=1 pour contourner)"
  flutter pub get
  flutter test
else
  echo "→ SKIP_TESTS=1 — tests Flutter ignorés"
fi

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
echo "  Privacy policy URL : ${PETSFOLLOW_SITE_STAGING}/legal/privacy"
echo "  Support contact : support@petsfollow.app"
echo "  Site prod (quand actif) : ${PETSFOLLOW_SITE_PROD}"
echo "  Checklist : documentation/26-PLAY-STORE.md"
echo "  Canal testeurs staging : make firebase-android-dist"
