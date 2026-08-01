#!/usr/bin/env bash
# Build APK release petsFollow pets (flavor staging) + upload Firebase App Distribution.
# Prérequis : firebase login, flutter, Android SDK.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FLUTTER_DIR="${ROOT}/flutter"
PROJECT_ID="${GCP_PROJECT_ID:-premedica-prod-2025}"
# App Firebase « petsFollow pets (staging) » — package be.llitsc.petsfollow_mobile.staging
ANDROID_APP_ID="${FIREBASE_ANDROID_APP_ID:-1:237481297060:android:a15f2666e4707cf1c9d231}"
# Domaine custom (api.petsfollow.ll-it-sc.be) : NXDOMAIN tant que DNS OVH non posé.
# URL Cloud Run directe (voir documentation/10-GCP-DEPLOIEMENT.md).
API_BASE="${API_BASE:-https://petsfollow-api-a7ako2njea-od.a.run.app}"
GOOGLE_SERVER_CLIENT_ID="${GOOGLE_SERVER_CLIENT_ID:-237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com}"
# Staging App Distribution: Client AI (CR explain + triage) on by default — tag dev.
CLIENT_AI_ENABLED="${CLIENT_AI_ENABLED:-true}"
GROUP_ALIAS="${APP_DIST_GROUP:-petsfollow-testers}"
RELEASE_NOTES="${RELEASE_NOTES:-petsFollow pets staging — Google Sign-In client (API ${API_BASE})}"

cd "${FLUTTER_DIR}"

if [[ "${SKIP_TESTS:-}" != "1" ]]; then
  echo "→ flutter test (SKIP_TESTS=1 pour contourner)"
  flutter pub get
  flutter test
else
  echo "→ SKIP_TESTS=1 — tests Flutter ignorés"
fi

PUBSPEC="${FLUTTER_DIR}/pubspec.yaml"
# Toujours bump le build number (+N) avant dist — requis App Distribution / versionCode.
CURRENT="$(grep -E '^version:' "${PUBSPEC}" | head -1 | awk '{print $2}')"
NAME="${CURRENT%%+*}"
BUILD="${CURRENT##*+}"
if [[ "${CURRENT}" != *"+"* ]] || ! [[ "${BUILD}" =~ ^[0-9]+$ ]]; then
  echo "version pubspec invalide: ${CURRENT} (attendu name+build, ex. 0.2.9+15)" >&2
  exit 1
fi
NEW_BUILD=$((BUILD + 1))
NEW_VERSION="${NAME}+${NEW_BUILD}"
sed -i "s/^version: .*/version: ${NEW_VERSION}/" "${PUBSPEC}"
echo "→ Version ${CURRENT} → ${NEW_VERSION}"
if [[ "${RELEASE_NOTES}" == "petsFollow pets staging — Google Sign-In client (API ${API_BASE})" ]]; then
  RELEASE_NOTES="petsFollow pets staging ${NEW_VERSION} (API ${API_BASE})"
fi

echo "→ flutter pub get"
flutter pub get

echo "→ Build APK release staging (API_BASE=${API_BASE})"
flutter build apk --release --flavor staging \
  --dart-define="FLAVOR=staging" \
  --dart-define="APP_ENV=staging" \
  --dart-define="API_BASE=${API_BASE}" \
  --dart-define="GOOGLE_SERVER_CLIENT_ID=${GOOGLE_SERVER_CLIENT_ID}" \
  --dart-define="CLIENT_AI_ENABLED=${CLIENT_AI_ENABLED}"

APK_PATH="${FLUTTER_DIR}/build/app/outputs/flutter-apk/app-staging-release.apk"
if [[ ! -f "${APK_PATH}" ]]; then
  # Fallback Gradle flavor path
  APK_PATH="${FLUTTER_DIR}/build/app/outputs/apk/staging/release/app-staging-release.apk"
fi
test -f "${APK_PATH}"

echo "→ Upload App Distribution (staging) → group ${GROUP_ALIAS}"
firebase appdistribution:distribute "${APK_PATH}" \
  --app "${ANDROID_APP_ID}" \
  --project "${PROJECT_ID}" \
  --groups "${GROUP_ALIAS}" \
  --release-notes "${RELEASE_NOTES}"

echo ""
echo "✓ Distribué staging ${NEW_VERSION}. Les testeurs ${GROUP_ALIAS} reçoivent un email / notif App Tester."
echo "  Package : be.llitsc.petsfollow_mobile.staging"
echo "  Console : https://console.firebase.google.com/project/${PROJECT_ID}/appdistribution"
echo "  N'oublie pas de committer flutter/pubspec.yaml (version ${NEW_VERSION})."
