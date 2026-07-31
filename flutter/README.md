# petsFollow pets — Flutter

App mobile client (face **pets**). Auth centralisée via l'API Go + PostgreSQL — **pas Firebase Auth**.

## Flavors Android (staging / prod)

| Flavor | Package | Canal | Commande |
|--------|---------|-------|----------|
| **staging** (défaut) | `be.llitsc.petsfollow_mobile.staging` | Firebase App Distribution (testeurs) | `make flutter-dev` · `make firebase-android-dist` |
| **prod** | `be.llitsc.petsfollow_mobile` | Google Play (futur) | `API_BASE=https://… make play-android-bundle` |

- Runtime : toujours **`--flavor X` + `--dart-define=FLAVOR=X` + `--dart-define=APP_ENV=X`** (même valeur). `AppEnv.validate()` refuse un mismatch ou un release Android sans define.
- Bandeau **STAGING** si canal staging.
- `google-services.json` : `android/app/src/staging/` et `android/app/src/prod/`.
- iOS : pas encore de flavors (bundle `be.llitsc.petsfollowMobile` inchangé).

### Testeurs — migration package

1. Installer **petsFollow Staging** via App Tester (nouvelle app, package `.staging`).
2. Désinstaller l’ancienne `petsFollow` (`be.llitsc.petsfollow_mobile`) pour éviter le chooser sur les deep links `petsfollow://`.
3. Si Google Sign-In échoue en release : enregistrer le SHA de la clé utilisée (debug et/ou upload) :
   `make firebase-google-signin-android SHA1=… FLAVOR=staging`

Deep links Stripe (même scheme sur staging et prod) : `petsfollow://payment/success` · `petsfollow://payment/cancel`

## Firebase

| Plateforme | Identifiant | App Firebase |
|------------|-------------|--------------|
| Android staging | `be.llitsc.petsfollow_mobile.staging` | `petsFollow pets (staging)` · id `…:a15f2666e4707cf1c9d231` |
| Android prod | `be.llitsc.petsfollow_mobile` | `petsFollow pets` |
| iOS | `be.llitsc.petsfollowMobile` | `petsFollow pets (iOS)` |

Projet GCP/Firebase : `premedica-prod-2025` (même que l'API staging).  
App Distribution : groupe `petsfollow-testers` doit être lié à l’app **staging** (pas l’ancienne app prod package).

Services Firebase utilisés : `firebase_core` + FCM.  
Connexion utilisateur : `POST /api/v1/auth/login` → JWT (`ApiClient` / secure storage).

Recréer / resynchroniser la config :

```bash
make firebase-flutter-setup
# ou
bash infra/firebase/setup-flutter-firebase.sh
```

## Lancer en local (staging)

```bash
# API déjà up (make api-dev) + GOOGLE_OAUTH_CLIENT_ID dans .env
make flutter-dev
# équivalent :
# cd flutter && flutter run --flavor staging \
#   --dart-define=FLAVOR=staging \
#   --dart-define=APP_ENV=staging \
#   --dart-define=API_BASE=http://10.0.2.2:8291 \
#   --dart-define=GOOGLE_SERVER_CLIENT_ID=237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com
```

Sans `GOOGLE_SERVER_CLIENT_ID`, le bouton Google est masqué.  
Prérequis Android : SHA debug/upload enregistrés sur l’app **staging** :
`make firebase-google-signin-android SHA1=… FLAVOR=staging` — voir [`documentation/26-PLAY-STORE.md`](../documentation/26-PLAY-STORE.md) §7.

## Firebase App Distribution (Android staging)

Groupe testeurs : `petsfollow-testers`  
API cible par défaut : Cloud Run `https://petsfollow-api-a7ako2njea-od.a.run.app`  
Package : `be.llitsc.petsfollow_mobile.staging`

```bash
make firebase-android-dist
# ou URL custom :
# API_BASE=https://… make firebase-android-dist
```

Console : https://console.firebase.google.com/project/premedica-prod-2025/appdistribution  
App Android Tester pour installer les builds invités.

## Google Play (Android App Bundle — flavor prod)

Prérequis :

1. Générer un upload keystore (une seule fois) :
   ```bash
   keytool -genkey -v -keystore flutter/android/upload-keystore.jks \
     -keyalg RSA -keysize 2048 -validity 10000 -alias upload
   ```
2. Copier `flutter/android/key.properties.example` → `flutter/android/key.properties` et renseigner mots de passe / alias (`storeFile=upload-keystore.jks` par défaut, relatif à `flutter/android/`).
3. Enregistrer les SHA-1/256 (upload + Play App Signing) dans Firebase / Google Cloud OAuth Android (**package prod**).

Build AAB :

```bash
API_BASE=https://… make play-android-bundle
# → flutter/build/app/outputs/bundle/prodRelease/app-prod-release.aab
```

`API_BASE` HTTPS est **obligatoire** (pas de défaut silencieux vers le Cloud Run staging).

Privacy policy (Play Console) : https://petsfollow.ll-it-sc.be/legal/privacy  
Checklist complète : [`documentation/26-PLAY-STORE.md`](../documentation/26-PLAY-STORE.md)
