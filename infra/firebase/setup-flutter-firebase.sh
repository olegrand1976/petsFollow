#!/usr/bin/env bash
# Enregistre l'app Flutter petsFollow sur Firebase (projet GCP partagé).
# Auth : PostgreSQL via API Go — Firebase Auth volontairement NON utilisé.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FLUTTER_DIR="${ROOT}/flutter"
PROJECT_ID="${GCP_PROJECT_ID:-premedica-prod-2025}"
ANDROID_PACKAGE="be.llitsc.petsfollow_mobile"
IOS_BUNDLE_ID="be.llitsc.petsfollowMobile"

echo "→ Projet Firebase/GCP : ${PROJECT_ID}"
echo "→ Android : ${ANDROID_PACKAGE}"
echo "→ iOS     : ${IOS_BUNDLE_ID}"

cd "${FLUTTER_DIR}"

if [[ ! -d android/app ]]; then
  echo "→ Génération plateformes android/ios"
  flutter create . --org be.llitsc --project-name petsfollow_mobile --platforms=android,ios
fi

echo "→ Création apps Firebase (idempotent — échoue si déjà existantes)"
firebase apps:create android "petsFollow pets" \
  --package-name="${ANDROID_PACKAGE}" \
  --project="${PROJECT_ID}" 2>/dev/null || true

firebase apps:create ios "petsFollow pets (iOS)" \
  --bundle-id="${IOS_BUNDLE_ID}" \
  --project="${PROJECT_ID}" 2>/dev/null || true

echo "→ FlutterFire configure"
flutterfire configure \
  --project="${PROJECT_ID}" \
  --platforms=android,ios \
  --android-package-name="${ANDROID_PACKAGE}" \
  --ios-bundle-id="${IOS_BUNDLE_ID}" \
  --yes \
  --out=lib/firebase_options.dart

echo "→ Dépendance firebase_core"
flutter pub add firebase_core

echo ""
echo "✓ Firebase Flutter prêt."
echo "  Auth : POST /api/v1/auth/login → JWT (PostgreSQL)"
echo "  Firebase : FCM / infra mobile uniquement — pas Firebase Auth"
echo ""
echo "Prochaines étapes manuelles (console Firebase) :"
echo "  1. Vérifier que Authentication n'a AUCUN provider activé pour cette app"
echo "  2. Activer Cloud Messaging quand FCM sera branché côté API Go"
