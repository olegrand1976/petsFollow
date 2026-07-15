# Stripe billing — petsFollow

Abonnement **par animal** via Stripe Checkout (paiement unique ou abonnement auto-renouvelé).

## Offres

| Plan | Code | Prix | Durée |
|------|------|------|-------|
| 25 € / an | `annual` | 2500 ct | 1 an |
| 60 € / 3 ans | `triennial` | 6000 ct | 3 ans (recommandé) |
| 75 € / 5 ans | `quinquennial` | 7500 ct | 5 ans |

Modes : `one_time` (Checkout `payment`) ou `subscription` (Checkout `subscription`, interval 1/3/5 ans).

## Variables d'environnement

```env
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ANNUAL_ONETIME=price_...
STRIPE_PRICE_TRIENNIAL_ONETIME=price_...
STRIPE_PRICE_QUINQUENNIAL_ONETIME=price_...
STRIPE_PRICE_ANNUAL_SUB=price_...
STRIPE_PRICE_TRIENNIAL_SUB=price_...
STRIPE_PRICE_QUINQUENNIAL_SUB=price_...
STRIPE_SUCCESS_URL=petsfollow://payment/success
STRIPE_CANCEL_URL=petsfollow://payment/cancel
BILLING_MOCK_ENABLED=true   # dev sans clé Stripe
PETSFOLLOW_API_PUBLIC_URL=http://localhost:8291
```

## Flux création animal (Flutter)

1. `POST /api/v1/pets` avec `plan`, `billingMode`, champs animal
2. Réponse `{ pet, checkoutUrl, sessionId }`
3. Redirection Stripe Checkout
4. Webhook `checkout.session.completed` → entitlement `active`
5. Renouvellements subscription via `invoice.paid`

## Webhooks traités

- `checkout.session.completed`
- `invoice.paid` / `invoice.payment_failed`
- `customer.subscription.updated` / `customer.subscription.deleted`

## Dev mock

Sans clé Stripe, `BILLING_MOCK_ENABLED=true` : le `checkoutUrl` pointe vers  
`GET /api/v1/billing/dev/mock-complete?pet_id=...&owner_user_id=...&plan_code=...&billing_mode=...`

## Admin plateforme

- `GET /admin/metrics/overview` — CA, MRR, conversion
- `GET /admin/users` — inscriptions
- `GET /admin/payments` — paiements reçus

Compte seed : `admin.demo@petsfollow.test` / `AdminDemo123!`
