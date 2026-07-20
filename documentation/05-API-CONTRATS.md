# Contrats API — petsFollow

Base : `http://localhost:8291/api/v1` (staging : `https://api.petsfollow.ll-it-sc.be/api/v1`).

## Conventions

- Succès : enveloppe `{ "data": ... }` (`httpx.WriteData`).
- Erreurs : code HTTP + corps erreur i18n (`code` / clé message).
- Auth : `Authorization: Bearer <access>` ; locale via `Accept-Language` + profil user.
- Côté Nuxt : BFF proxy tel quel ; pages : `const items = res.data ?? res`.

## Groupes de routes

| Groupe | Exemples (préfixe `/api/v1`) |
|--------|------------------------------|
| Auth public | `POST /auth/login`, `/register`, `/confirm-email`, `/forgot-password`, `/reset-password`, `/refresh`, `/auth/google`, `/auth/2fa/verify` |
| Journey public | `GET/POST /public/journey/unsubscribe?token=` (opt-out parcours email) |
| Auth protégé | `GET/POST /auth/2fa/*` |
| Me | `GET/PATCH /me`, avatar, password, locale, vets, household, discovery, device-tokens |
| Véto | `/clients`, `/vet/*` (profile, availability messagerie, overview, link-requests, prospects, commissions, prefs) |
| Calendrier RDV | `GET/PUT /vet/schedule`, `GET/POST/DELETE /vet/vacations`, `GET /vet/calendar`, `GET /practices/{id}/availability`, `GET/POST /pets/{id}/visits`, `PATCH /visits/{id}` (`confirm` / `propose_reschedule` / `accept_reschedule` / `reject_reschedule` / `cancel`) |
| Pets / FC | `/pets`, heartrate sessions, timeline, photo, care-reminders, visits, horse-* |
| Messaging | `/messaging/threads…` |
| Billing | `GET /billing/plans`, `/billing/addons`, webhook Stripe, checkout/portal pet, my-addons |
| Commercial | `/commercial/overview`, `/vets`, `/prospects`, `/commissions`, `GET/PATCH /commercial/me/payout-profile` |
| Admin | `/admin/metrics/overview`, `/users`, `/payments`, `/commercials`, `/prospects`, `/vets`, `/clients` |
| Admin import clients | `POST/GET /admin/client-imports`, `GET …/{id}`, `POST …/suggest-mapping`, `PUT …/mapping`, `PATCH …/rows/{rowId}`, `POST …/commit`, `GET …/credentials` — voir [24](24-IMPORT-CLIENTS-ADMIN.md) |
| Admin commissions véto | `GET /admin/commissions/runs`, `GET …/periods/{YYYY-MM}`, `POST …/close`, `POST …/mark-paid`, `PUT /admin/commissions/tiers`, `GET/PUT /admin/commissions/settings` (PUT rejette : taux commercial = constantes plan) |
| Admin commissions commercial | `GET /admin/commercial-commissions/runs`, `GET …/periods/{YYYY-MM}`, `POST …/close`, `POST …/mark-paid` |
| Admin SPIFF | `GET /admin/commercial-bonuses`, `POST /admin/commercial-bonuses/{id}/mark-paid` |

Handlers : `go/internal/handlers/` (`api.go`, `auth.go`, `billing.go`, `admin.go`, `commercial.go`, `commissions.go`, …).

## Billing webhook

Endpoint code : `POST /api/v1/billing/webhooks/stripe` (signing secret `STRIPE_WEBHOOK_SECRET`).

## Pas de catalogue exhaustif

Pour le détail request/response, lire les handlers + tests d’intégration (`*_integration_test.go`) et smoke (`make smoke`).
