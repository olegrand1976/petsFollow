# AGENTS.md — petsFollow

## Projet

Monorepo **petsFollow** : continuité de soins vétérinaire — **Pro** (Web) · **Pro Light** (mobile terrain) · **app client** (mobile particulier).

| Face | Stack | Port dev |
|------|-------|----------|
| **Pro** (véto/admin) | Nuxt 3 (`nuxtjs/`) | **3002** |
| **pets** (clients) | Flutter (`flutter/`) | — |
| **API** | Go (`go/`) | **8291** |

## Démarrage local

```bash
# Terminal 1 — infra + API
make up-infra && make migrate && make seed && make api-dev

# Terminal 2 — Nuxt Pro
make nuxtjs-dev   # http://localhost:3002
```

Après modification des tokens brand : `make brand-sync`.

## Comptes démo

Mot de passe commun véto : `VetDemo123!` · client : `ClientDemo123!` · admin / DEV : `AdminDemo123!` · commercial : `CommercialDemo123!` · care_pro : `CareProDemo123!` · research : `ResearchDemo123!`

| Rôle | Email | Cabinet |
|------|-------|---------|
| Véto | `vet.demo@petsfollow.test` | Cabinet VetPlus Demo — multi-profil aussi `research` (switch → `/research`) |
| Véto | `vet.parc@petsfollow.test` | Clinique du Parc |
| Véto | `vet.lyon@petsfollow.test` | Centre Cardio Animaux Lyon |
| Véto | `vet.onboarding@petsfollow.test` | Onboarding (profil incomplet) |
| Véto | `vet.unverified@petsfollow.test` | Email non confirmé |
| Véto | `vet.reset@petsfollow.test` | Token démo reset MDP |
| Care pro (farrier) | `farrier.demo@petsfollow.test` | Flutter pro light — Spirit (write_notes) · dual profil client |
| Care pro (vet_light) | `vetlight.demo@petsfollow.test` | Flutter pro light — Spirit · dual profil client |
| Équipe VetPlus | `vet.colleague@` · `vet.assist@` · `secretary.demo@` | MDP `VetDemo123!` — page `/team` |
| Commercial manager | `commercial.manager@petsfollow.test` | Bérénice — équipe démo forcée au seed : Camille + Alex (2 reps) |
| Commercial | `commercial.demo@petsfollow.test` | Camille — vet.demo assigné, 5 prospects Bruxelles, rattaché manager |
| Commercial | `commercial.demo2@petsfollow.test` | Alex — vet.parc assigné, 5 prospects Nord, rattaché manager |

Entraînement pitch IA (commercial) : `/commercial/training` — nécessite `GEMINI_API_KEY`. Admin : `/admin/training`. Module CR IA (essai 90j / 39 € HT) : `/admin/ai-modules`, `/commercial/ai-modules`, guide `/commercial/ai-cr-playbook` — job quotidien adhésion+friction : `POST /api/v1/internal/ai-module-friction/run` + `X-Ai-Module-Friction-Secret` (`AI_MODULE_FRICTION_SECRET`, scheduler `infra/gcp/setup-ai-module-friction-scheduler.sh`). Analyseur quotidien : `POST /api/v1/internal/pitch-analyzer/run` + header `X-Pitch-Analyzer-Secret`. Purge RGPD 3 ans d'inactivité : `POST /api/v1/internal/retention/run` + header `X-Retention-Secret` (env `RETENTION_PURGE_SECRET`, scheduler quotidien `make gcp-retention-scheduler`) — purge aussi les partages de dossier périmés (lignes + ZIP en bucket) et annule les walk-ins `consultation_session` orphelins (`confirmed`, sans CR persisté, `scheduled_at` ≥ 6 h ; filet au prochain run quotidien). Brouillons SaaS Flux A (C1) : `POST /api/v1/internal/saas-invoices/run` + `X-Saas-Invoices-Secret` (`SAAS_INVOICES_SECRET`, scheduler mensuel `make gcp-saas-invoices-scheduler`) — draft only (boucle batches, opt-in `saas_billing_enabled`, mois Europe/Brussels), send Peppol manuel admin `/admin/invoicing`. Smoke live master : `make billit-saas-master-smoke`. Auto-branches commerciaux : `POST /api/v1/internal/sales-branches-auto/run` + `X-Sales-Branches-Auto-Secret` (`SALES_BRANCHES_AUTO_SECRET`, scheduler 10h/18h `make gcp-sales-branches-scheduler`) — crée une branche `DUPONTD` pour les commerciaux sans branche et non rattachés à un peer, + email de félicitations. Pharmacie péremption : `POST /api/v1/internal/pharmacy/expiry-run` + `X-Pharmacy-Expiry-Secret` (`PHARMACY_EXPIRY_SECRET`, scheduler quotidien 04:00 `make gcp-pharmacy-expiry-scheduler`) — auto-quarantaine + digest email le lundi. Digest produit quotidien (admin/commercial) : ingest GH Action + `POST /api/v1/internal/product-digest/run` à 18:00 Brussels — voir `documentation/25-PRODUCT-DIGEST.md`.
| Admin | `admin.demo@petsfollow.test` | — (global) — multi-profils `admin` + `client` + `vet` VetPlus |
| DEV (support IT) | `dev.demo@petsfollow.test` | Ops sans billing/seed — multi-profils `dev` + `client` + `vet` VetPlus |
| Research (épidémio) | `research.demo@petsfollow.test` | Observatoire `/research` (tag `dev`) — aussi profil `research` sur `vet.demo` / `admin.demo` |
| Client (Flutter) | `client.demo@petsfollow.test` | VetPlus — 6 pets démo + Care+/Kennel/Horse (seed) · tél. `0470 00 00 01` |
| Client | `client.vide@petsfollow.test` | VetPlus — sans animal |
| Client | `client.marie@petsfollow.test` | Parc — Mimi, Chouchou |
| Client | `client.paul@petsfollow.test` | Parc — Max |
| Client | `client.julie@petsfollow.test` | Lyon — Oscar |
| Client | `client.thomas@petsfollow.test` | Lyon — Luna, Nico (pending) |

Confirmation email démo : `/confirm-email?token=demo-confirm-email` 
Reset mot de passe démo : `/reset-password?token=demo-reset-password` (`vet.reset@petsfollow.test`)

Médias (avatars / photos) : local = `./data/uploads` servi sous `/media/` ; staging = bucket GCS `petsfollow-media` (`make gcp-setup-media`, env `GCS_MEDIA_BUCKET`). PHI CR (`visit-reports/`), carnets (`health-books/`) et packs de partage (`dossier-shares/`) : pas d’URL publique (stream auth) — GCP refuse une IAM conditionnelle sur `allUsers`. Lifecycle GCS : `dossier-shares/**` supprimé au-delà de 2 jours.

Relancer les données : `make seed`

**Densification démo (prod-like)** : après le seed de base, `make seed-mass` ajoute des comptes `mass.*@petsfollow.test` (≈20 cabinets/vétos, ≈360 clients, ≈700 animaux, 8 care_pros), rattache les nouveaux vétos aux commerciaux Camille/Alex (inchangés), et recalcule les commissions (`AccrueAll*`). Idempotent (skip si déjà présent). Pour régénérer : `make seed && make seed-mass`.

**Garde-fou seed** : `seed` **et** `seed-mass` refusent de tourner si `APP_ENV` n'est pas dans l'allowlist `dev` / `development` / `local` / `test` / `staging` — une variable absente ou mal orthographiée bloque au lieu de laisser passer. `APP_ENV` est posé par les cibles Make (défaut `local`) et par `pf_write_api_env_file` côté Cloud Run (défaut **`production`**, `staging` passé explicitement en 4e argument par `cloudbuild.yaml` et `setup-jobs.sh`).

**Staging GCP** : pas de seed auto (Scheduler supprimé : `make gcp-delete-seed-scheduler`). Reset manuel : admin Pro (zone danger, phrase `RESET STAGING`) ou `bash infra/gcp/postdeploy.sh --seed`. Annonce staff après seed : email auto / commande `seed-notify`. Au reset : tickets support conservés ; client `b.murgo1976@gmail.com` + graphe propriétaire préservés (remap cabinets démo).

## Tests

**Philosophie** : toute mutation métier = test au niveau le plus bas possible (Go intégration > Playwright `@p0` > widget Flutter). Non effet de bord (billing mock, users jetables, seed `*.petsfollow.test`). Règle Cursor : `.cursor/rules/anti-regression-quality.mdc`. Checklist P0/P1 + auto : [`documentation/15-PLAN-TESTS.md`](documentation/15-PLAN-TESTS.md).

**Use cases commerciaux** : scénarios manuels non-tech → dossier [`useCase/`](useCase/) ; page Pro staging `/usecases` (admin / commercial / manager, layout selon rôle). Badge **S** topbar si `NUXT_PUBLIC_APP_ENV=staging` + session auth. Sync catalogue : `make usecases-sync` / garde-fou `make usecases-check` (CI). Règle `.cursor/rules/usecase-sync.mdc`.

**Modules Pro tag `dev`** (badge `nav.tagDev`, pas GA) : facturation `/invoicing` · pharmacie `/medicaments` `/stock` `/daf` (`PHARMACY_ENABLED`) · **AFMPS/Compendium** `/admin/afmps-imports` `/admin/compendium-imports` (même flag, doc [`27-PHARMACIE-BELGIQUE.md`](documentation/27-PHARMACIE-BELGIQUE.md)) · **prescriptions** `/prescriptions` (`PRESCRIPTIONS_ENABLED`, doc [`35-PRESCRIPTIONS.md`](documentation/35-PRESCRIPTIONS.md)) · **PACS** onglet Imagerie + `/admin/pacs` (`PACS_ENABLED`, doc [`40-PACS.md`](documentation/40-PACS.md) — local : `make up-pacs` + `make pacs-demo-seed`) · **Research** `/research` + `/admin/research` (`RESEARCH_ENABLED`, doc [`42-RESEARCH.md`](documentation/42-RESEARCH.md) — scheduler `make gcp-research-etl-scheduler`). Règle `.cursor/rules/modules-tag-dev.mdc`.

```bash
# Unitaires + intégration Go (intégration skip si DB absente ; sinon make up-infra)
make test-go

# Unitaires Nuxt (Vitest)
make test-nuxt   # ou: cd nuxtjs && npm test

# Flutter widget/unit (+ smoke API opt-in)
make test-flutter
make test-flutter-smoke   # API :8291 seedée

# E2E Playwright P0 (API :8291 + Nuxt :3002 + seed requis)
make test-e2e-p0
# Suite complète : make test-e2e

# Smoke API (login + H1 messagerie croisée + HR comment + timeline + register flows)
make smoke
```

`make test` = Go + Nuxt + Flutter. Prérequis e2e : `make up-infra && make migrate && make seed`, API et `make nuxtjs-dev` démarrés. `make api-dev` fixe `AUTH_RATE_LIMIT_PER_MIN=1000` par défaut (prod : 60/min par IP) — sans ça, la suite e2e complète déclenche des 429 sur `/auth/*`.

## Sécurité & RGPD

- **Auth Nuxt** : tokens JWT en cookies **httpOnly** posés par la BFF (`pf_token`, `pf_refresh`) + marqueur `pf_session` lisible client. Logout : `POST /api/auth/logout` (BFF). Token pour WebSocket : `GET /api/auth/ws-token`. **Ne jamais** exposer un JWT au JS client (localStorage, cookie non-httpOnly).
- **Auth Flutter** : tokens dans `flutter_secure_storage` (migration one-shot depuis shared_preferences).
- **Rate limit** : endpoints `/auth/*` publics limités par IP — `AUTH_RATE_LIMIT_PER_MIN` (60/min prod, 1000 en dev via `make api-dev`, 0 = off). La BFF propage `X-Forwarded-For`, honoré par `middleware.RealIP` côté Go.
- **Headers** : CSP + HSTS via `routeRules` (`nuxtjs/nuxt.config.ts`, helper `buildCsp()`) ; côté Go CORS allowlist (`CORS_ALLOWED_ORIGINS`) + `X-Content-Type-Options`/`X-Frame-Options`/`Referrer-Policy`/HSTS. Fonts auto-hébergées (`nuxtjs/assets/css/fonts.css` + `public/fonts/`) — pas de CDN Google Fonts.
- **Secrets** : `JWT_SIGNING_KEY` obligatoire hors dev (fail-fast au boot). Les secrets d'endpoints internes (`X-Retention-Secret`, `X-Pitch-Analyzer-Secret`, `X-Product-Digest-Secret`, `X-Ai-Module-Friction-Secret`, `X-Sales-Branches-Auto-Secret`) se comparent en temps constant via `secretHeaderOK` (`go/internal/handlers/retention.go`) — utiliser ce helper pour tout nouvel endpoint interne.
- **RGPD** : doc [`documentation/36-RGPD.md`](documentation/36-RGPD.md) · export `GET /api/v1/me/export` (JSON, BFF `me/export.get.ts`, bouton `/settings` + `/commercial/settings` + profil Flutter) · suppression `DELETE /me` (purge complète client / anonymisation pro tombstone ; purge client-owned avant tombstone dual profil) · consentement obligatoire au register (`"consent": true` → `identity.users.terms_accepted_at`) · clients provisionnés : `POST /me/accept-terms` + gate Flutter si `termsAcceptedAt` null · purge auto 3 ans d'inactivité (`last_login_at`, cron `internal/retention/run`) · pré-dialogue avant permission push Flutter.
- **Dev only** : `confirmPath`/`resetPath` dans les réponses register/forgot/resend-confirmation exposés uniquement si `DEV_SEED_ENABLED=true` (les e2e/smoke en dépendent). `BILLING_MOCK_ENABLED` opt-in explicite (défaut true seulement via `make api-dev`) ; signature webhook Stripe toujours vérifiée, même en mock.
- **Resend confirm** : `POST /auth/resend-confirmation` `{email}` (rate-limité, toujours 200) — BFF Nuxt `/api/auth/resend-confirmation` · Flutter login si `email_not_verified` · Pro `/register/sent`.
- **Partage de dossier animal** : le client Flutter envoie le dossier complet (PHI) à un pro par mail après consentement explicite dans le dialogue. Le lien public `/dossier/{token}` (Nuxt, hors auth, `Referrer-Policy: no-referrer` + `X-Robots-Tag: noindex`) expire à 24 h, est rate-limité côté Go (anti-énumération), plafonné à 10 envois/jour/propriétaire, et les lignes + ZIP périmés sont purgés par le job de rétention et par la lifecycle GCS.
- **Téléphone commercial** : `COMMERCIAL_CONTACT_PHONE` (API Go) est un **repli global optionnel** sur la page de dossier public. Vide en staging/prod est **volontaire** (pas un oubli de deploy) : la source réelle est le `contact_phone` du commercial rattaché, garanti par la porte `/complete-contact-phone` et le seed. Les rôles `commercial` / `commercial_manager` y sont redirigés tant qu'ils n'ont pas renseigné le leur (middleware global Nuxt, `PATCH /api/v1/me`).
- **Alertes auth ALERT/URGENT** : échec SMTP confirm / spikes login / clients non vérifiés → ticket `source=system` dans `/admin/support` + email `OPS_NOTIFY_EMAIL` (staging : `o.legrand1976@gmail.com`). Job horaire `POST /internal/auth-health/run` + `X-Auth-Health-Secret` (`AUTH_HEALTH_SECRET`, script `infra/gcp/setup-auth-health-scheduler.sh`).
- **Purge RGPD (one-shot ops)** : après premier deploy du câblage secret, exécuter `RETENTION_PURGE_SECRET=… make gcp-retention-scheduler` puis **redeploy** l'API (sinon `pf_api_secrets` ignore le secret absent et `/internal/retention/run` reste en 401). Lifecycle filet : `make gcp-setup-media` (préfixe `dossier-shares/`, 2 jours, **merge** avec les règles existantes). Même convention manuelle pour auth-health / ai-module-friction / saas-invoices / sales-branches-auto / pharmacy-expiry / research-etl (`SAAS_INVOICES_SECRET=… make gcp-saas-invoices-scheduler`, `SALES_BRANCHES_AUTO_SECRET=… make gcp-sales-branches-scheduler`, `PHARMACY_EXPIRY_SECRET=… make gcp-pharmacy-expiry-scheduler`, `RESEARCH_ETL_SECRET=… RESEARCH_ANON_SALT=… make gcp-research-etl-scheduler` puis `--update-secrets` ou redeploy). Le secret du job Scheduler est en clair dans la config HTTP du job : restreindre `scheduler.jobs.get` ; le script n'affiche plus le YAML du job.
- **2FA** : anti-replay TOTP (`totpReplayGuard`) — un code ne passe qu'une fois par fenêtre.

## Langues (FR / NL / EN / ES / ET / IT)

Locales supportées : `fr` (défaut), `nl`, `en`, `es`, `et`, `it`.

| Face | Mécanisme | Persistance |
|------|-----------|-------------|
| **Nuxt Pro** | `@nuxtjs/i18n`, cookie `pf_locale` | `PATCH /api/v1/me/locale` via `/settings` |
| **Flutter** | `gen-l10n` + `LocaleController` | `shared_preferences` + `PATCH /me/locale` |
| **API Go** | middleware `Accept-Language` + `users.preferred_locale` | emails/billing/erreurs traduits |

Compte démo NL : `client.marie@petsfollow.test` (`preferred_locale = nl`).

Après migration : `make migrate` (000005_user_locale, 000018_locale_es, 000040_locale_et, 000063_locale_it).

## Google OAuth + 2FA (optionnel)

| Variable | Où | Description |
|----------|-----|-------------|
| `GOOGLE_OAUTH_CLIENT_ID` | API Go | Client ID Google Web (validation idToken) |
| `NUXT_PUBLIC_GOOGLE_CLIENT_ID` | Nuxt | Même Client ID (bouton Google sur `/login`) |
| `GOOGLE_SERVER_CLIENT_ID` | Flutter (`--dart-define`) | Même Client ID Web (`google_sign_in` → idToken) |

Sans ces variables, la connexion email/mot de passe fonctionne normalement ; le bouton Google est masqué.

**Flutter pets** : Google Sign-In avec `audience=client` — lie un compte client existant **ou crée** un compte client si l’email est inconnu (create-if-absent). Un email Pro → erreur `google_client_only`. Self-signup email : `POST /auth/register-client` (écran inscription Flutter). `POST /auth/register` et `/auth/register-client` exigent `"consent": true` (CGU/privacy, horodaté dans `identity.users.terms_accepted_at`).

**Flutter care_pro** : même app, shell pro light (`ProLightShell`) après login `role=care_pro` + `professionalSpecialty`. Création via admin (`POST /admin/care-pros`) ; register public off par défaut (`CARE_PRO_PUBLIC_REGISTER`).

**2FA** : activation dans Paramètres (`/settings`) — TOTP via application authenticator.

## UI Pro (Nuxt)

- Design system : composants `Pro*` dans `nuxtjs/components/pro/`
- Logo : `components/PetsFollowLogo.vue` (variants default/compact/hero)
- Shell : `ProSidebar` + `ProTopbar` (notifs véto uniquement)
- Listes : `ProListToolbar` + bascule table/kanban (`useListView`)
- CSS : `nuxtjs/assets/css/pro-*.css` + tokens `--pf-vet-*`
- Règle Cursor : `.cursor/rules/petsfollow-pro-ui.mdc`
- Charte : `documentation/13-CHARTE-GRAPHIQUE.md`

**Ne pas** mélanger le thème dark Flutter dans Pro.

## API

- Base : `http://localhost:8291/api/v1`
- Réponses enveloppées `{ data: ... }` — BFF Nuxt proxy tel quel
- Côté pages : `const items = res.data ?? res`

## Firebase (Flutter pets + push API)

- Projet : `premedica-prod-2025` (GCP partagé)
- Apps Android : **staging** `be.llitsc.petsfollow_mobile.staging` (App Distribution, `make firebase-android-dist`) · **prod** `be.llitsc.petsfollow_mobile` (Play, `make play-android-bundle`) · iOS `be.llitsc.petsfollowMobile`
- Dev local Flutter : `make flutter-dev` → flavor **staging** + `FLAVOR`/`APP_ENV=staging` + bandeau STAGING
- **Auth** : PostgreSQL via API Go (`/api/v1/auth/login`) — **ne pas** activer Firebase Auth
- Setup apps : `make firebase-flutter-setup`
- **Push FCM** : l’API Go envoie les push (message véto → client, confirmation RDV) via ADC (`GOOGLE_APPLICATION_CREDENTIALS` ou SA Cloud Run). Désactiver : `FCM_ENABLED=false`. Détails : `documentation/08-MESSAGERIE-NOTIFICATIONS.md`.

## Structure clé

- `go/internal/handlers/` — routes API
- `nuxtjs/pages/` — pages Pro (véto + admin)
- `nuxtjs/server/api/` — BFF Nuxt
- `brand/tokens/design-tokens.json` — source tokens
