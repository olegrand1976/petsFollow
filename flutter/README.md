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
