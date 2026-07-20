# Stripe billing — petsFollow

Abonnement **par animal** via Stripe Checkout (paiement unique ou abonnement auto-renouvelé).  
Addons foyer (Family / Kennel / Care+ / Horse) : **paiement unique à vie** (`Checkout` `payment`, `valid_until` NULL).

## Offres

> Politique économique complète : [`17-POLITIQUE-TARIFAIRE.md`](./17-POLITIQUE-TARIFAIRE.md).

| Plan | Code | Prix | Durée entitlement | Modes Stripe |
|------|------|------|-------------------|--------------|
| 35 € / an | `annual` | 3500 ct | 1 an | `one_time` **et** `subscription` (`year`×1) |
| 95 € / 3 ans | `triennial` | 9500 ct | 3 ans (recommandé) | `one_time` **et** `subscription` (`year`×3) |
| 145 € / 5 ans | `quinquennial` | 14500 ct | 5 ans | **`one_time` uniquement** (Stripe refuse un intervalle récurrent > 3 ans) |

Modes : `one_time` (Checkout `payment`) ou `subscription` (Checkout `subscription`, interval max **3 ans**).

## Variables d'environnement

```env
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ANNUAL_ONETIME=price_...
STRIPE_PRICE_TRIENNIAL_ONETIME=price_...
STRIPE_PRICE_QUINQUENNIAL_ONETIME=price_...
STRIPE_PRICE_ANNUAL_SUB=price_...
STRIPE_PRICE_TRIENNIAL_SUB=price_...
# STRIPE_PRICE_QUINQUENNIAL_SUB — non utilisé (Stripe max 3 ans en récurrent)
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
- [ ] Créer **5 Prices** abos (annual / triennial × one_time+sub · quinquennial **one_time only**) + **4 Prices addons one-time** (Family / Kennel / Care+ / Horse)
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
| **Family pack** | `family` | 39 € (une fois) | owner (≥2 animaux ; vue foyer ; remise plans **−10 %**) |
| **Kennel pack** | `kennel` | 119 € (une fois) | owner (≥6 animaux ; batch / `litter_tag` ; remise plans **−15 %**) |
| **Care+** | `care_plus` | 19 € (une fois) | owner (médicaments / rappels perso ; export & emails = roadmap) |
| **Horse pack** | `horse` | 39 € (une fois) | owner (pets `horse` : maréchal, contacts, compétitions) |

Mode Stripe : Checkout **`payment`** (one-time) pour les 4 addons. Activation → `valid_until` **NULL** (à vie). Pas de renouvellement.  
Colonne `billing.addon_entitlements.stripe_subscription_id` (migration `000023`) : **legacy** uniquement (anciens abos yearly) ; handlers `invoice.paid` / `subscription.*` restent pour ces lignes.

Cycle de vie :
- Reject post-paiement (éligibilité Family/Kennel/Care+/Horse déjà actif) → entitlement DB cancelled (+ cancel Stripe sub si legacy `subID`)
- Upgrade Kennel → deactivate Family DB (+ cancel sub Family legacy si présent)
- Nouveaux achats : `payment` + lifetime ; anti-rachet si addon déjà actif/pending
- **Ops migration** : pour clients encore en sub addon, cancel Stripe sub + `UPDATE valid_until = NULL` (manuelle)

API : `GET /billing/addons`, `POST /billing/addons/checkout`, `GET /billing/my-addons` (client). Webhooks : `checkout.session.completed` (`metadata.kind=addon`) · legacy `invoice.paid` / `payment_failed` / `customer.subscription.*` si `stripe_subscription_id` non vide.

Variables Stripe : `STRIPE_PRICE_ADDON_FAMILY`, `STRIPE_PRICE_ADDON_KENNEL`, `STRIPE_PRICE_ADDON_CARE_PLUS`, `STRIPE_PRICE_ADDON_HORSE` (Prices **one-time**).

Customer Portal : surtout pour les abonnements animal (les addons one-time n’ont plus de sub à gérer).

Commissions (assiette **HTVA du montant payé**, TVA BE 21 % ; Prices Stripe = **TTC**) :

| Offre | Commercial | Véto |
|-------|------------|------|
| Annual | 8 % | progressif × facteur |
| Triennial | **12 %** | progressif × facteur |
| Quinquennial | 8 % | progressif × facteur |
| Family / Kennel | **10 %** | **5 %** |
| Care+ / Horse | **10 %** | **0 %** |

Détail économique → [17-POLITIQUE-TARIFAIRE.md](./17-POLITIQUE-TARIFAIRE.md) · fiches → [18](./18-FICHE-COMMISSION-VETO.md) / [19](./19-FICHE-COMMISSION-COMMERCIAL.md).

### Mise à jour Prices Stripe (ops)

Après bascule tarifaire, recréer les Prices Live/Test et mettre à jour les secrets :

| Secret Manager | Variable |
|----------------|----------|
| `petsfollow-stripe-price-annual-onetime` | `STRIPE_PRICE_ANNUAL_ONETIME` |
| `petsfollow-stripe-price-triennial-onetime` | `STRIPE_PRICE_TRIENNIAL_ONETIME` |
| `petsfollow-stripe-price-quinquennial-onetime` | `STRIPE_PRICE_QUINQUENNIAL_ONETIME` |
| `petsfollow-stripe-price-annual-sub` | `STRIPE_PRICE_ANNUAL_SUB` |
| `petsfollow-stripe-price-triennial-sub` | `STRIPE_PRICE_TRIENNIAL_SUB` |
| `petsfollow-stripe-price-addon-family` | `STRIPE_PRICE_ADDON_FAMILY` |
| `petsfollow-stripe-price-addon-kennel` | `STRIPE_PRICE_ADDON_KENNEL` |
| `petsfollow-stripe-price-addon-care-plus` | `STRIPE_PRICE_ADDON_CARE_PLUS` |
| `petsfollow-stripe-price-addon-horse` | `STRIPE_PRICE_ADDON_HORSE` |

Montants attendus : 35 / 95 / 145 € (abos) ; 39 / 119 / 19 / 39 € **one-time** (addons Family / Kennel / Care+ / Horse).
