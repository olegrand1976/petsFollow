# 26 — Google Play Store (Android petsFollow)

Checklist pour un passage **complet** de l’app Flutter **prod** `be.llitsc.petsfollow_mobile` sur Google Play.

Le canal **testeurs** est distinct : flavor **staging** (`be.llitsc.petsfollow_mobile.staging`) via Firebase App Distribution — voir [`flutter/README.md`](../flutter/README.md).

Privacy policy URL (obligatoire listing) : **https://petsfollow.ll-it-sc.be/legal/privacy**  
CGU : https://petsfollow.ll-it-sc.be/legal/terms  
Mentions : https://petsfollow.ll-it-sc.be/legal/mentions  
Support : **support@ll-it-sc.be**

---

## 1. Prérequis techniques (repo)

| Item | Statut attendu |
|------|----------------|
| Flavor Play | `prod` (`applicationId` sans suffixe) |
| Flavor testeurs | `staging` (suffixe `.staging`) — **pas** sur Play Production |
| `targetSdk` / `compileSdk` | 36 (Flutter) |
| `minSdk` | 24 |
| Signing release | `flutter/android/key.properties` + `upload-keystore.jks` (gitignored) |
| AAB | `API_BASE=https://… make play-android-bundle` |
| Dart defines | `--dart-define=FLAVOR=prod` **et** `APP_ENV=prod` (mêmes valeurs ; `AppEnv.validate`) |
| `API_BASE` release | `https://…` obligatoire (pas de défaut staging) |
| `allowBackup` | `false` + règles XML |
| `AD_ID` | retiré (`tools:node="remove"`) — déclarer « Non » Advertising ID |
| Suppression compte | in-app (Profil) + API `DELETE /api/v1/me` |

### Générer le keystore upload

```bash
keytool -genkey -v -keystore flutter/android/upload-keystore.jks \
  -keyalg RSA -keysize 2048 -validity 10000 -alias upload
cp flutter/android/key.properties.example flutter/android/key.properties
# Éditer storePassword / keyPassword / keyAlias (storeFile=upload-keystore.jks par défaut)
```

Afficher les empreintes à coller dans Firebase / Cloud Console :

```bash
keytool -list -v -keystore flutter/android/upload-keystore.jks -alias upload
```

Après le **premier** upload Play, récupérer aussi le certificat **App Signing by Google Play** (Play Console → App integrity) et l’ajouter comme SHA Firebase / client OAuth Android (**package prod**).

### Build AAB (prod)

```bash
API_BASE=https://… make play-android-bundle
# → flutter/build/app/outputs/bundle/prodRelease/app-prod-release.aab
```

Firebase App Distribution (APK **staging** testeurs) : `make firebase-android-dist`.

---

## 2. Play Console — créer l’app

1. Compte développeur Google Play (organisation LL-IT-SC).
2. Créer l’application : package `be.llitsc.petsfollow_mobile`, type App, gratuit / payant selon modèle (abonnements via Stripe hors Play Billing).
3. Activer **Play App Signing**.
4. Uploader le premier AAB en piste **Internal testing** (puis Closed / Production).

### Compte développeur personnel (si applicable)

Google peut exiger un **Closed testing** (≥ 12 testeurs, ≥ 14 jours) avant Production. Prévoir la piste fermée dès le premier AAB.

---

## 3. Fiche Play (Store listing)

| Champ | Contenu suggéré |
|-------|-----------------|
| Nom | petsFollow |
| Description courte | Suivi santé prescrit avec votre vétérinaire — app mobile |
| Description longue | Messagerie cabinet, rappels de soins, relevés respiratoires, foyer, abonnement animal (mensuel / annuel / triennal)… |
| Icône 512×512 | Exporter depuis `brand/` |
| Feature graphic 1024×500 | Visuel marketing |
| Screenshots téléphone | ≥ 2 (login, home, relevé, soins, messagerie) |
| Catégorie | Médical / Santé & fitness (animaux) |
| Email contact | support@ll-it-sc.be |
| Privacy policy | https://petsfollow.ll-it-sc.be/legal/privacy |
| Site | https://petsfollow.ll-it-sc.be |

Localiser FR / EN / NL / ES si disponible dans Console.

---

## 4. Contenu & audience

| Déclaration | Réponse |
|-------------|---------|
| Designed for Families / kids | **Non** — adultes (propriétaires) |
| Target age | 18+ |
| News app / COVID | Non |
| Contenu généré utilisateur | Messages / photos — modération cabinet |

Remplir le **questionnaire IARC** (Content rating).

---

## 5. Data Safety (formulaire Console)

Déclarer la **collecte** (et le partage avec sous-traitants le cas échéant) :

| Donnée | Collectée | Partagée | Finalité |
|--------|-----------|----------|----------|
| Nom / email | Oui | Non (sauf Google Sign-In) | Compte |
| Photos / vidéos | Oui | Cabinet (médias) | Profil animal, messagerie |
| Messages | Oui | Cabinet | Messagerie |
| Santé / FR (animal) | Oui | Cabinet | Suivi santé (dont relevés respiratoires) |
| Identifiants appareil / FCM | Oui | Google FCM | Notifications |
| Infos paiement | Via Stripe | Stripe | Abonnements animal (plans) |
| Advertising ID | **Non** | — | — |

Préciser : chiffrement en transit (HTTPS), suppression possible (in-app + email support), conservation (compte / 3 ans inactivité).

### Permissions sensibles (déclaration Photo / Vidéo)

- `CAMERA` : photo animal / pièce jointe
- `READ_MEDIA_IMAGES` / `READ_MEDIA_VIDEO` : galerie avatars & messagerie  
Justifier comme fonctionnalités **cœur** (pas one-time).

---

## 6. Suppression de compte (politique Play)

- **In-app** : Profil → Supprimer le compte (déjà livré).
- **Web / email** : indiquer sur la privacy policy et éventuellement une page dédiée ; contact `support@ll-it-sc.be` pour demande hors app.
- Dans Play Console → App content → Account deletion : décrire le chemin in-app + URL privacy / email.

---

## 7. OAuth Google Sign-In

1. Client OAuth **Android** :
   - **staging** : `be.llitsc.petsfollow_mobile.staging` (SHA debug + upload)
   - **prod** : `be.llitsc.petsfollow_mobile` (SHA upload + App Signing Play)
2. Enregistrer les empreintes puis régénérer le `google-services.json` du flavor :
   ```bash
   make firebase-google-signin-android SHA1=ED:B0:… SHA256=7A:67:… FLAVOR=staging
   make firebase-google-signin-android SHA1=ED:B0:… SHA256=7A:67:… FLAVOR=prod
   ```
3. Conserver le Web client ID (`GOOGLE_SERVER_CLIENT_ID` / `GOOGLE_OAUTH_CLIENT_ID`) pour `idToken` serveur :
   `237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com`
4. API locale / Cloud Run : `GOOGLE_OAUTH_CLIENT_ID` = ce Web client ID. Flutter staging : `make flutter-dev` (flavor + dart-define inclus).

Sans client Android (`client_type: 1`) dans le `google-services.json` du flavor, Play Services renvoie `DEVELOPER_ERROR` (code 10).

---

## 8. Déclarations financières

Paiements **Stripe Checkout** (hors Google Play Billing) pour abonnements animaux (monthly / annual / triennial).  
Dans App content : déclarer les achats numériques / abonnements selon le questionnaire Play (paiements externes / liens).

---

## 9. Parcours de publication recommandé

```text
1. keystore + key.properties + SHA Firebase (package prod)
2. API_BASE=https://… make play-android-bundle   # flavor prod
3. Upload AAB → Internal testing (smoke)
4. Closed testing (si requis) + notes de version
5. Remplir Data Safety, Content rating, Account deletion, Ads ID = No
6. Store listing + screenshots + privacy URL
7. Production → Review Google
```

Canal testeurs (indépendant de Play) : `make firebase-android-dist` (flavor staging).

---

## 10. Vérifications post-build

- [ ] AAB signé avec l’alias `upload` (pas debug)
- [ ] App lance avec API HTTPS staging/prod
- [ ] Google Sign-In fonctionne avec SHA Play
- [ ] Lien « Voir en ligne » privacy ouvre https://petsfollow.ll-it-sc.be/legal/privacy
- [ ] Suppression de compte fonctionne
- [ ] Pas de demande de permission AD_ID
- [ ] Notifications (POST_NOTIFICATIONS) demandées au runtime
