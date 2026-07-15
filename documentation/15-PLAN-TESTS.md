# Plan de tests — petsFollow

## API (smoke)

`make smoke` — health, auth véto/client/admin, clients, billing mock, messagerie, heartrate validate, timeline.

## Web Pro (Playwright)

Répertoire : `nuxtjs/tests/e2e/specs/`

| Spec | Scénario |
|------|----------|
| `01-auth` | Login véto → dashboard → clients |
| `02-locale` | Changement langue EN dans settings |
| `03-clients` | Recherche client |
| `04-messaging` | Page messagerie + deep-link thread |
| `05-onboarding` | Redirection véto profil incomplet |
| `06-admin` | Login admin → dashboard |

Local : `cd nuxtjs && npm run test:e2e` (API + Nuxt sur 8291/3002).

CI PR : `npx playwright test --list` (validation des specs).

Staging : workflow `deploy-gcp-staging.yml` exécute Playwright contre `petsfollow.ll-it-sc.be`.

## Go / Flutter

- Go : `make test-go`
- Nuxt unit : `make test-nuxt` (Vitest)
- Flutter : `make test-flutter`
