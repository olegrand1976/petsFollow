# petsFollow pets — Flutter

App mobile client (face **pets**). Auth centralisée via l'API Go + PostgreSQL — **pas Firebase Auth**.

## Firebase

| Plateforme | Identifiant | App Firebase |
|------------|-------------|--------------|
| Android | `be.llitsc.petsfollow_mobile` | `petsFollow pets` |
| iOS | `be.llitsc.petsfollowMobile` | `petsFollow pets (iOS)` |

Projet GCP/Firebase : `premedica-prod-2025` (même que l'API staging).

Services Firebase utilisés : `firebase_core` (base pour FCM futur).  
Connexion utilisateur : `POST /api/v1/auth/login` → JWT stocké localement (`ApiClient`).

Recréer / resynchroniser la config :

```bash
make firebase-flutter-setup
# ou
bash infra/firebase/setup-flutter-firebase.sh
```

Fichiers générés : `lib/firebase_options.dart`, `android/app/google-services.json`, `ios/Runner/GoogleService-Info.plist`.

## Lancer en local

```bash
# API déjà up (make api-dev)
cd flutter
flutter pub get
flutter run --dart-define=API_BASE=http://10.0.2.2:8291   # émulateur Android
# flutter run --dart-define=API_BASE=http://localhost:8291  # iOS simulateur / device
```

Deep links Stripe : `petsfollow://payment/success` · `petsfollow://payment/cancel`

## Firebase App Distribution (Android)

Groupe testeurs : `petsfollow-testers`  
API cible par défaut : Cloud Run `https://petsfollow-api-a7ako2njea-od.a.run.app`  
(domaine `api.petsfollow.ll-it-sc.be` : à activer via DNS OVH — voir doc GCP)

```bash
make firebase-android-dist
# ou URL custom :
# API_BASE=https://… make firebase-android-dist
```

Console : https://console.firebase.google.com/project/premedica-prod-2025/appdistribution  
App Android Tester pour installer les builds invités.

## Google Play (Android App Bundle)

Prérequis :

1. Générer un upload keystore (une seule fois) :
   ```bash
   keytool -genkey -v -keystore flutter/android/upload-keystore.jks \
     -keyalg RSA -keysize 2048 -validity 10000 -alias upload
   ```
2. Copier `flutter/android/key.properties.example` → `flutter/android/key.properties` et renseigner mots de passe / alias (`storeFile=upload-keystore.jks` par défaut, relatif à `flutter/android/`).
3. Enregistrer les SHA-1/256 (upload + Play App Signing) dans Firebase / Google Cloud OAuth Android.

Build AAB :

```bash
make play-android-bundle
# → flutter/build/app/outputs/bundle/release/app-release.aab
```

Privacy policy (Play Console) : https://petsfollow.ll-it-sc.be/legal/privacy  
Checklist complète : [`documentation/26-PLAY-STORE.md`](../documentation/26-PLAY-STORE.md)
