#!/usr/bin/env bash
# Build APK release petsFollow pets + upload Firebase App Distribution.
# Prérequis : firebase login, flutter, Android SDK.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FLUTTER_DIR="${ROOT}/flutter"
PROJECT_ID="${GCP_PROJECT_ID:-premedica-prod-2025}"
ANDROID_APP_ID="${FIREBASE_ANDROID_APP_ID:-1:237481297060:android:cfda5c59a08bfd6dc9d231}"
# Domaine custom (api.petsfollow.ll-it-sc.be) : NXDOMAIN tant que DNS OVH non posé.
# URL Cloud Run directe (voir documentation/10-GCP-DEPLOIEMENT.md).
API_BASE="${API_BASE:-https://petsfollow-api-a7ako2njea-od.a.run.app}"
GROUP_ALIAS="${APP_DIST_GROUP:-petsfollow-testers}"
RELEASE_NOTES="${RELEASE_NOTES:-petsFollow pets — build staging (API ${API_BASE})}"

cd "${FLUTTER_DIR}"

echo "→ flutter pub get"
flutter pub get

echo "→ Build APK release (API_BASE=${API_BASE})"
flutter build apk --release --dart-define="API_BASE=${API_BASE}"

APK_PATH="${FLUTTER_DIR}/build/app/outputs/flutter-apk/app-release.apk"
test -f "${APK_PATH}"

echo "→ Upload App Distribution → group ${GROUP_ALIAS}"
firebase appdistribution:distribute "${APK_PATH}" \
  --app "${ANDROID_APP_ID}" \
  --project "${PROJECT_ID}" \
  --groups "${GROUP_ALIAS}" \
  --release-notes "${RELEASE_NOTES}"

echo ""
echo "✓ Distribué. Les testeurs ${GROUP_ALIAS} reçoivent un email / notif App Tester."
echo "  Console : https://console.firebase.google.com/project/${PROJECT_ID}/appdistribution"
