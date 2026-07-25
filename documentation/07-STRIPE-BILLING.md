# Stripe billing — petsFollow

Abonnement **par animal** via Stripe Checkout (paiement unique ou abonnement auto-renouvelé).  
Anciens addons (Family / Kennel / Care+ / Horse) et quinquennial : **plus en vente** (legacy code / secrets éventuels).

## Offres

> Politique économique complète : [`17-POLITIQUE-TARIFAIRE.md`](./17-POLITIQUE-TARIFAIRE.md).

| Plan | Code | Prix | Durée entitlement | Modes Stripe |
|------|------|------|-------------------|--------------|
| 3,50 € / mois | `monthly` | 350 ct | 1 mois | **`subscription` only** (`month`×1) |
| 35 € / an | `annual` | 3500 ct | 1 an | `one_time` **et** `subscription` (`year`×1) |
| 95 € / 3 ans | `triennial` | 9500 ct | 3 ans (recommandé) | `one_time` **et** `subscription` (`year`×3) |

**Deprecated / hors vente** : `quinquennial` (145 € / 5 ans, one_time legacy) · checkout addons.

Modes : `one_time` (Checkout `payment`) ou `subscription` (Checkout `subscription`, interval max **3 ans**). Monthly = sub only.

**Entitlement animal actif** ouvre Care / Horse / foyer / kennel encoding. HR premium + messagerie restent gated sur l’entitlement payant.

## Variables d'environnement

```env
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
# Fallback si pas de mapping actif en DB (billing.stripe_prices) :
STRIPE_PRICE_MONTHLY_SUB=price_...
STRIPE_PRICE_ANNUAL_ONETIME=price_...
STRIPE_PRICE_TRIENNIAL_ONETIME=price_...
STRIPE_PRICE_ANNUAL_SUB=price_...
STRIPE_PRICE_TRIENNIAL_SUB=price_...
# Legacy (ne plus créer / ne plus vendre) :
# STRIPE_PRICE_QUINQUENNIAL_ONETIME
# STRIPE_PRICE_ADDON_FAMILY / KENNEL / CARE_PLUS / HORSE
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
- `GET /admin/stripe-catalog` — catalogue produits/prix Stripe (DB)
- `POST /admin/stripe-catalog/import` — import CSV Dashboard (`kind=products|prices`, multipart `file`) — upsert

Compte seed : `admin.demo@petsfollow.test` / `AdminDemo123!`

### Catalogue Stripe (DB)

Tables `billing.stripe_products` / `billing.stripe_prices` (migration `000046`, **schéma seul** — pas d’IDs Stripe dans la migration).

Le checkout résout le Price ID ainsi :

1. Price **active** mappée `(plan_code, billing_mode)` en DB
2. Sinon fallback env `STRIPE_PRICE_*` (Secret Manager)
3. Erreur DB catalogue → checkout échoue (pas de fallback silencieux)

**Charger le catalogue :**

| Environnement | Méthode |
|---------------|---------|
| Local / démo | `make seed` (upsert IDs de référence dans `store.UpsertDefaultStripeCatalog`) |
| Staging / Live | Admin UI `/admin/stripe-catalog` — import CSV Dashboard (`products.csv` puis `prices.csv`) |

Import admin : montants EU (`3,50`) acceptés ; entier sans décimale = cents (style `unit_amount`). Mapping plan/mode inféré depuis intervalle Stripe ou nom produit.

Montants commerciaux (`domain.go`) restent la source d’offre / commissions ; le Price Stripe détermine le débit hors remise foyer.
---

## Mise en production (checklist)

### 1. Dashboard Stripe (mode Live)

- [ ] Activer le compte Stripe en mode **Live**
- [ ] Créer **5 Prices** vendables : monthly **sub** · annual / triennial × one_time+sub (pas de quin, pas d’addons)
- [ ] Noter chaque `price_…` ID pour les variables `STRIPE_PRICE_*`
- [ ] Activer le **Customer Portal** (Settings → Billing → Customer portal) pour la gestion d'abonnement Flutter

### 2. Webhook

- [ ] Endpoint : `https://api.petsfollow.ll-it-sc.be/api/v1/billing/webhooks/stripe` (ou URL Cloud Run API)
- [ ] Événements : `checkout.session.completed`, `invoice.paid`, `invoice.payment_failed`, `customer.subscription.updated`, `customer.subscription.deleted`
- [ ] Copier le signing secret `whsec_…`

### 3. Secrets GCP (Secret Manager)

```bash
./infra/gcp/setup-stripe-secrets.sh
# Puis remplacer les placeholders REPLACE_ME :
# echo -n 'sk_live_...' | gcloud secrets versions add petsfollow-stripe-secret-key --data-file=-
```

Secrets attendus (offre actuelle) :

| Secret Manager | Variable Cloud Run |
|----------------|-------------------|
| `petsfollow-stripe-secret-key` | `STRIPE_SECRET_KEY` |
| `petsfollow-stripe-webhook-secret` | `STRIPE_WEBHOOK_SECRET` |
| `petsfollow-stripe-price-monthly-sub` | `STRIPE_PRICE_MONTHLY_SUB` |
| `petsfollow-stripe-price-annual-onetime` | `STRIPE_PRICE_ANNUAL_ONETIME` |
| `petsfollow-stripe-price-triennial-onetime` | `STRIPE_PRICE_TRIENNIAL_ONETIME` |
| `petsfollow-stripe-price-annual-sub` | `STRIPE_PRICE_ANNUAL_SUB` |
| `petsfollow-stripe-price-triennial-sub` | `STRIPE_PRICE_TRIENNIAL_SUB` |

Montants attendus : **3,50 / 35 / 95 €**. Legacy quin / addons : ne plus provisionner.

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

## Legacy addons / quinquennial

Code et tables `addon_entitlements` peuvent encore exister pour l’historique. **Checkout addons et vente quinquennial sont dépréciés.**  
Handlers webhook legacy (`metadata.kind=addon`, `stripe_subscription_id` addon) restent pour les lignes existantes.

Commissions actuelles (assiette **HTVA**, TVA BE 21 %) :

| Offre | Commercial | Véto |
|-------|------------|------|
| Monthly | 8 % | progressif × 0,67 |
| Annual | 8 % | progressif × 0,67 |
| Triennial | **12 %** | progressif × 1,00 |

Détail → [17-POLITIQUE-TARIFAIRE.md](./17-POLITIQUE-TARIFAIRE.md) · fiches → [18](./18-FICHE-COMMISSION-VETO.md) / [19](./19-FICHE-COMMISSION-COMMERCIAL.md).
