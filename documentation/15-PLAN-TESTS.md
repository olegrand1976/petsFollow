# Plan de tests — petsFollow

## API (smoke)

`make smoke` — health, auth véto/client/admin, clients, billing mock, messagerie, heartrate validate **avec comment**, timeline.

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
| `07-commercial` | Login commercial → overview / prospects |
| `08-commercial-manager` | Dashboard équipe |
| `08-requests` | Calendrier + invitations clients |
| `09-pet-detail` | Fiche animal, shares, commentaire relevé HR |
| `10-products` | `/produits` plans TTC 3,50 / 35 / 95 |
| `11-admin-stripe-catalog` | Catalogue Stripe admin + ACL véto |

Local : `cd nuxtjs && npm run test:e2e` (API + Nuxt sur 8291/3002).

CI PR : `npx playwright test --list` (validation des specs).

Staging : workflow `deploy-gcp-staging.yml` exécute Playwright contre `petsfollow.ll-it-sc.be`.

## Go / Flutter

- Go unit + intégration : `make test-go` (heartrate comment, stripe catalog, care_pro ACL, multi-profils)
- Nuxt unit : `make test-nuxt` (Vitest)
- Flutter unit/widget : `make test-flutter` (palette, comment HR payload, pro light nav)
