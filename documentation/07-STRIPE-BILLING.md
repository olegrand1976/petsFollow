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

---

## Mise en production (checklist)

### 1. Dashboard Stripe (mode Live)

- [ ] Activer le compte Stripe en mode **Live**
- [ ] Créer 6 **Prices** (annual / triennial / quinquennial × `one_time` + `subscription`)
- [ ] Noter chaque `price_…` ID pour les variables `STRIPE_PRICE_*`
- [ ] Activer le **Customer Portal** (Settings → Billing → Customer portal) pour la gestion d'abonnement Flutter

### 2. Webhook

- [ ] Endpoint : `https://api.petsfollow.ll-it-sc.be/api/v1/billing/webhook` (ou URL Cloud Run API)
- [ ] Événements : `checkout.session.completed`, `invoice.paid`, `invoice.payment_failed`, `customer.subscription.updated`, `customer.subscription.deleted`
- [ ] Copier le signing secret `whsec_…`

### 3. Secrets GCP (Secret Manager)

```bash
./infra/gcp/setup-stripe-secrets.sh
# Puis remplacer les placeholders REPLACE_ME :
# echo -n 'sk_live_...' | gcloud secrets versions add petsfollow-stripe-secret-key --data-file=-
```

Secrets attendus :

| Secret Manager | Variable Cloud Run |
|----------------|-------------------|
| `petsfollow-stripe-secret-key` | `STRIPE_SECRET_KEY` |
| `petsfollow-stripe-webhook-secret` | `STRIPE_WEBHOOK_SECRET` |
| `petsfollow-stripe-price-annual-onetime` | `STRIPE_PRICE_ANNUAL_ONETIME` |
| `petsfollow-stripe-price-triennial-onetime` | `STRIPE_PRICE_TRIENNIAL_ONETIME` |
| `petsfollow-stripe-price-quinquennial-onetime` | `STRIPE_PRICE_QUINQUENNIAL_ONETIME` |
| `petsfollow-stripe-price-annual-sub` | `STRIPE_PRICE_ANNUAL_SUB` |
| `petsfollow-stripe-price-triennial-sub` | `STRIPE_PRICE_TRIENNIAL_SUB` |
| `petsfollow-stripe-price-quinquennial-sub` | `STRIPE_PRICE_QUINQUENNIAL_SUB` |

Monter les secrets sur Cloud Run via `--set-secrets` (fusionner avec `pf_api_secrets` dans `infra/gcp/lib/deploy-run-args.sh`).

### 4. Déploiement

- [ ] `BILLING_MOCK_ENABLED=false` — contrôlé par variable d'environnement au deploy :

```bash
export BILLING_MOCK_ENABLED=false
# cloudbuild ou deploy manuel Cloud Run
```

Par défaut, `infra/gcp/lib/deploy-run-args.sh` utilise `BILLING_MOCK_ENABLED="${BILLING_MOCK_ENABLED:-true}"` pour éviter un billing live accidentel.

- [ ] Vérifier `STRIPE_SUCCESS_URL` / `STRIPE_CANCEL_URL` (deep links Flutter)
- [ ] Smoke test : création animal → Checkout live → webhook → entitlement `active`

### 5. Post-go-live

- [ ] Surveiller les logs webhook API
- [ ] Tester le portail client (`POST /api/v1/billing/portal`)
- [ ] Vérifier métriques admin (`/admin/metrics/overview`)

---

## Monétisation addons

| Pack | Code | Prix | Scope |
|------|------|------|-------|
| **Family pack** | `family` | 40 € / an | owner (multi-animaux) |
| **Care+** | `care_plus` | 15 € / an | par animal |
| **Horse pack** | `horse` | 30 € / an | par animal |

API : `GET /billing/addons`, `POST /billing/addons/checkout` (client). Webhook `checkout.session.completed` avec `metadata.kind=addon`.

Variables Stripe optionnelles : `STRIPE_PRICE_ADDON_FAMILY`, `STRIPE_PRICE_ADDON_CARE_PLUS`, `STRIPE_PRICE_ADDON_HORSE`.

Commission **commercial** : **miroir** de la commission du véto assigné (même `rate_bps` / `commission_cents` à l’activation) + **15 %** sur addons Family / Care+ / Horse.
