# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Vue d'ensemble

Monorepo **petsFollow** — continuité de soins vétérinaire, trois faces :

| Face | Stack | Dossier | Port dev |
|------|-------|---------|----------|
| **Pro** (véto / admin / commercial) | Nuxt 3 | `nuxtjs/` | 3002 |
| **pets** (clients) + **Pro Light** (care_pro terrain) | Flutter | `flutter/` | — |
| **API** | Go + Chi | `go/` | 8291 (`/api/v1`) |

**[AGENTS.md](AGENTS.md)** (déjà utilisé par Cursor) est la référence d'onboarding : comptes démo seed, modules tag `dev`, sécurité/RGPD détaillée, Firebase, jobs internes/schedulers. Le lire avant toute tâche impliquant des comptes de test ou un module flaggé. Les règles Cursor dans `.cursor/rules/*.mdc` s'appliquent aussi à Claude Code (résumé ci-dessous).

## Commandes

### Démarrage local (2 terminaux)

```bash
# T1 — infra (db + redis + mailhog) + API
make up-infra && make migrate && make seed && make api-dev   # API :8291

# T2 — Web Pro
make nuxtjs-dev    # http://localhost:3002

# Optionnel
make flutter-dev   # Flutter flavor staging (émulateur ; device physique : API_BASE=http://<LAN>:8291)
make up-pacs && make pacs-demo-seed   # Orthanc local :8042 (démo Imagerie)
```

`make api-dev` force les flags dev : `DEV_SEED_ENABLED`, `BILLING_MOCK_ENABLED`, `AUTH_RATE_LIMIT_PER_MIN=1000`, et active les modules tag `dev` (`PHARMACY_ENABLED`, `PRESCRIPTIONS_ENABLED`, `PACS_ENABLED`, `RESEARCH_ENABLED`, `CLIENT_AI_ENABLED`…). Relancer les données démo : `make seed` (puis `make seed-mass` pour densifier). Après modification des tokens brand : `make brand-sync`.

### Tests

```bash
make test-go            # Go unit + intégration (-p 1 : DB partagée ; intégration skip si DB absente)
make test-nuxt          # Vitest + usecases-check
make test-flutter       # widget/unit Flutter
make test-e2e-p0        # Playwright @p0 — nécessite API :8291 + Nuxt :3002 + seed
make test-e2e           # suite Playwright complète (rate limit à 1000 requis, cf. api-dev)
make test               # Go + Nuxt + Flutter
make smoke              # smoke API (login, messagerie, timeline, register)
make test-auth          # garde-fou login/forgot (Go + Vitest) avant deploy
```

Test unique :

```bash
cd go && go test ./internal/handlers/ -run 'TestNomDuTest' -count=1
cd nuxtjs && npm test -- tests/unit/monFichier.spec.ts
cd nuxtjs && npx playwright test tests/e2e/specs/03-clients.spec.ts
cd flutter && flutter test test/features/mon_test.dart
```

### Migrations

Fichiers SQL numérotés dans `go/internal/platform/db/migrations/` (`000NNN_nom.up.sql` / `.down.sql`). Appliquer : `make migrate`.

## Architecture

### Flux

Flutter appelle directement l'API Go. Nuxt passe par sa **BFF** (`nuxtjs/server/api/`) qui proxifie vers l'API Go — les réponses sont enveloppées `{ data: ... }`, d'où le pattern côté pages : `const items = res.data ?? res`. L'API s'appuie sur PostgreSQL (schémas `identity`, `practice`, `pets`, `heartrate`, `messaging`, `notifications`, `billing`, `sales`, `care`, `visits`, `discovery`), Redis, Stripe (mock en dev), et des médias en local `./data/uploads` → `/media/` (GCS en staging).

### Go (`go/internal/`)

- `handlers/` — routes API ; tests d'intégration via `newTestAPI` (`auth_integration_test.go`)
- `store/` — accès données ; `billing/` — Stripe + montants (`billing/domain.go` = source de vérité des prix)
- `platform/` — transversal : config, db/migrations, authx, i18n, httpx, gemini (IA), pdf…
- `seed/` — données démo ; `workers/`, `pharmacy/`, `prescription/`, `invoicing/` (Billit), `notifications/` (FCM)
- Point d'entrée : `cmd/petsfollow-api` (sous-commandes `migrate`, `seed`, `seed-mass`, `import-cnk`)

### Nuxt (`nuxtjs/`)

- `pages/` — pages Pro (véto + admin + commercial) ; `server/api/` — BFF
- Design system : composants `Pro*` dans `components/pro/` (`ProCard`, `ProTable`, `ProListToolbar` + bascule table/kanban via `useListView`, …), icônes **Material Symbols** via `ProIcon` (auto-hébergées — jamais de CDN Google Fonts), tokens CSS `--pf-vet-*` / `--pf-brand-*` (`assets/css/tokens.css`, générés par `make brand-sync` depuis `brand/tokens/design-tokens.json`)
- Face **light** uniquement — ne pas importer le thème dark Flutter dans Pro
- Règle UI complète : `.cursor/rules/petsfollow-pro-ui.mdc` + `documentation/13-CHARTE-GRAPHIQUE.md`

### Flutter (`flutter/lib/`)

- `features/` par domaine, `core/`, `l10n/` (gen-l10n)
- Deux shells dans la même app : client (pets) et `ProLightShell` (care_pro)
- Auth = API Go (PostgreSQL) — **ne pas** activer Firebase Auth ; Firebase = FCM/App Distribution uniquement

### i18n — couverture par face

L'API accepte **8** locales, les faces n'en servent pas autant : ne pas raisonner sur un chiffre global.

| Face | Locales servies | Source de vérité |
|------|-----------------|------------------|
| **API Go** | `fr` `nl` `en` `es` `et` `it` `uk` `ru` | `i18n.Supported` (`platform/i18n/locale.go`) |
| **Flutter** | idem (les 8) | `LocaleController.supportedCodes` + les `app_*.arb` présents |
| **Nuxt Pro** | `fr` `nl` `en` `es` `et` `it` | `i18n.locales` (`nuxt.config.ts`) ⇄ `SUPPORTED_LOCALES` (`useLocaleSync.ts`) |

`fr` est le template partout. Règles :

- **Nuxt** : toute chaîne UI passe par `nuxtjs/locales/*.json` — **les 6 fichiers servis** à chaque ajout de clé. `nuxtjs/locales/{uk,ru}.json` existent mais sont **hors config** (traduction incomplète) : ne pas les réactiver sans `presentation/` et `ai-flows/` complets. `tests/unit/locales-parity.spec.ts` verrouille la parité.
- **Flutter** : `flutter gen-l10n` puis **committer** les `app_localizations*.dart` générés. `test/l10n/arb_parity_test.dart` verrouille parité, placeholders et cohérence `supportedCodes` ⇄ `AppLocalizations.supportedLocales`.
- **API Go** : middleware `Accept-Language` + `users.preferred_locale`. `TestCatalogParity` verrouille les catalogues.
- **Cyrillique** : DM Sans / IBM Plex Mono n'ont pas de glyphes cyrilliques → sous-sets Noto auto-hébergés (`nuxtjs/public/fonts/README.md`) ; PDF Go via `platform/pdffont` (Liberation Sans, métriquement compatible Arial) ; Flutter via `fontFamilyFallback`. SMS cyrilliques = UCS-2, **70 caractères** par segment (cf. `catalog_sms_test.go`).

### Modules tag `dev` (pas GA)

Facturation Billit, pharmacie/DAF, prescriptions, PACS, Research, Client AI : gardés derrière des flags (`*_ENABLED` API + `NUXT_PUBLIC_*_ENABLED` Nuxt), badge `nav.tagDev` en sidebar, endpoints 404 `*_disabled` si off. Ne retirer le badge qu'au passage GA explicite. Détail : `.cursor/rules/modules-tag-dev.mdc`.

## Règles à respecter (issues de `.cursor/rules/`)

### Anti-régression (`anti-regression-quality.mdc`)

Toute mutation métier = test **au niveau le plus bas possible** : intégration Go (`newTestAPI`) > Playwright `@p0` > widget Flutter. Même PR : code + test + maj `documentation/15-PLAN-TESTS.md` si parcours P0/P1 impacté. Non effet de bord : billing mock, users jetables pour les mutations destructives, comptes `*.petsfollow.test` uniquement. Ne pas contourner les gates CI (pas de skip intégration/Playwright, pas de deploy `gcp-deploy` nu — préférer push branche `staging`).

### Sécurité & RGPD (`securite-auth-rgpd.mdc`)

- **Jamais** de JWT accessible au JS client : cookies **httpOnly** (`pf_token`, `pf_refresh`) posés par la BFF (`nuxtjs/server/utils/api.ts`) ; Flutter : `flutter_secure_storage`.
- Nouvel endpoint auth public → rate limiter `authRL.Middleware` (`go/internal/handlers/api.go`). Register exige `"consent": true`.
- Secrets d'endpoints internes : comparaison temps constant via `secretHeaderOK` (`go/internal/handlers/retention.go`).
- Ressource externe côté Nuxt → l'ajouter explicitement dans `buildCsp()` (`nuxt.config.ts`) ; CORS Go = allowlist, pas de wildcard.
- Nouvelle table/colonne avec données personnelles → couvrir export (`store/user_export.go`), suppression/anonymisation (`handlers/retention.go`), et rétention 3 ans si applicable.

### Sync produit/commercial

- Changement d'offre/prix → maj `pages/produits.vue` + `index.productHighlights` dans **les 6 locales servies par Nuxt**, aligné sur `documentation/17-POLITIQUE-TARIFAIRE.md` et `billing/domain.go` (`produits-sync.mdc`).
- Nouveau parcours démo / changement d'écran / compte seed → maj `useCase/**/UC-*.md` + `make usecases-sync` (CI : `make usecases-check`) ; nouveau compte seed → aussi `AGENTS.md` (`usecase-sync.mdc`).

### Flutter (`flutter-action-tests.mdc`, `firebase-android-dist-version.mdc`)

- Action mutante dans `flutter/lib/features/` → `Key('feature_action')` stable + test widget sous `flutter/test/features/` + maj `flutter/test/COVERAGE_MATRIX.md`.
- Avant tout `make firebase-android-dist` : bump du build number dans `flutter/pubspec.yaml` (`+N` → `+(N+1)`), committé avec le deploy.
