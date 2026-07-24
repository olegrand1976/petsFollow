# AGENTS.md — petsFollow

## Projet

Monorepo **petsFollow** : suivi cardiaque vétérinaire dual-face.

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

Mot de passe commun véto : `VetDemo123!` · client : `ClientDemo123!` · admin : `AdminDemo123!` · commercial : `CommercialDemo123!` · care_pro : `CareProDemo123!`

| Rôle | Email | Cabinet |
|------|-------|---------|
| Véto | `vet.demo@petsfollow.test` | Cabinet VetPlus Demo |
| Véto | `vet.parc@petsfollow.test` | Clinique du Parc |
| Véto | `vet.lyon@petsfollow.test` | Centre Cardio Animaux Lyon |
| Véto | `vet.onboarding@petsfollow.test` | Onboarding (profil incomplet) |
| Véto | `vet.unverified@petsfollow.test` | Email non confirmé |
| Véto | `vet.reset@petsfollow.test` | Token démo reset MDP |
| Care pro (farrier) | `farrier.demo@petsfollow.test` | Flutter pro light — Spirit (write_notes) |
| Care pro (vet_light) | `vetlight.demo@petsfollow.test` | Flutter pro light — Spirit (write_notes) |
| Commercial manager | `commercial.manager@petsfollow.test` | Responsable commercial (équipe) |
| Commercial | `commercial.demo@petsfollow.test` | Force de vente (vet.demo assigné, rattaché manager) |

Entraînement pitch IA (commercial) : `/commercial/training` — nécessite `GEMINI_API_KEY`. Admin : `/admin/training`. Analyseur quotidien : `POST /api/v1/internal/pitch-analyzer/run` + header `X-Pitch-Analyzer-Secret`. Digest produit quotidien (admin/commercial) : ingest GH Action + `POST /api/v1/internal/product-digest/run` à 18:00 Brussels — voir `documentation/25-PRODUCT-DIGEST.md`.
| Admin | `admin.demo@petsfollow.test` | — (global) |
| Client (Flutter) | `client.demo@petsfollow.test` | VetPlus — 6 pets démo + Care+/Kennel/Horse (seed) |
| Client | `client.vide@petsfollow.test` | VetPlus — sans animal |
| Client | `client.marie@petsfollow.test` | Parc — Mimi, Chouchou |
| Client | `client.paul@petsfollow.test` | Parc — Max |
| Client | `client.julie@petsfollow.test` | Lyon — Oscar |
| Client | `client.thomas@petsfollow.test` | Lyon — Luna, Nico (pending) |

Confirmation email démo : `/confirm-email?token=demo-confirm-email` 
Reset mot de passe démo : `/reset-password?token=demo-reset-password` (`vet.reset@petsfollow.test`)

Médias (avatars / photos) : local = `./data/uploads` servi sous `/media/` ; staging = bucket GCS `petsfollow-media` (`make gcp-setup-media`, env `GCS_MEDIA_BUCKET`). PHI CR (`visit-reports/`) : pas d’URL publique (stream auth) — GCP refuse une IAM conditionnelle sur `allUsers`.

Relancer les données : `make seed`

## Tests

```bash
# Unitaires + intégration Go (intégration skip si DB absente ; sinon make up-infra)
make test-go

# Unitaires Nuxt (Vitest)
make test-nuxt   # ou: cd nuxtjs && npm test

# E2E Playwright auth (API :8291 + Nuxt :3002 + seed requis)
cd nuxtjs && npm run test:e2e -- tests/e2e/specs/01-auth.spec.ts
cd nuxtjs && npm run build

# Smoke API (login + register/confirm/forgot/reset)
make smoke
```

Prérequis e2e auth : `make up-infra && make migrate && make seed`, API et `make nuxtjs-dev` démarrés.

## Langues (FR / NL / EN / ES / ET)

Locales supportées : `fr` (défaut), `nl`, `en`, `es`, `et`.

| Face | Mécanisme | Persistance |
|------|-----------|-------------|
| **Nuxt Pro** | `@nuxtjs/i18n`, cookie `pf_locale` | `PATCH /api/v1/me/locale` via `/settings` |
| **Flutter** | `gen-l10n` + `LocaleController` | `shared_preferences` + `PATCH /me/locale` |
| **API Go** | middleware `Accept-Language` + `users.preferred_locale` | emails/billing/erreurs traduits |

Compte démo NL : `client.marie@petsfollow.test` (`preferred_locale = nl`).

Après migration : `make migrate` (000005_user_locale, 000018_locale_es, 000040_locale_et).

## Google OAuth + 2FA (optionnel)

| Variable | Où | Description |
|----------|-----|-------------|
| `GOOGLE_OAUTH_CLIENT_ID` | API Go | Client ID Google Web (validation idToken) |
| `NUXT_PUBLIC_GOOGLE_CLIENT_ID` | Nuxt | Même Client ID (bouton Google sur `/login`) |
| `GOOGLE_SERVER_CLIENT_ID` | Flutter (`--dart-define`) | Même Client ID Web (`google_sign_in` → idToken) |

Sans ces variables, la connexion email/mot de passe fonctionne normalement ; le bouton Google est masqué.

**Flutter pets** : Google Sign-In avec `audience=client` — lie un compte client existant **ou crée** un compte client si l’email est inconnu (create-if-absent). Un email Pro → erreur `google_client_only`. Self-signup email : `POST /auth/register-client` (écran inscription Flutter).

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
- Apps : Android `be.llitsc.petsfollow_mobile` · iOS `be.llitsc.petsfollowMobile`
- **Auth** : PostgreSQL via API Go (`/api/v1/auth/login`) — **ne pas** activer Firebase Auth
- Setup apps : `make firebase-flutter-setup`
- **Push FCM** : l’API Go envoie les push (message véto → client, confirmation RDV) via ADC (`GOOGLE_APPLICATION_CREDENTIALS` ou SA Cloud Run). Désactiver : `FCM_ENABLED=false`. Détails : `documentation/08-MESSAGERIE-NOTIFICATIONS.md`.

## Structure clé

- `go/internal/handlers/` — routes API
- `nuxtjs/pages/` — pages Pro (véto + admin)
- `nuxtjs/server/api/` — BFF Nuxt
- `brand/tokens/design-tokens.json` — source tokens
