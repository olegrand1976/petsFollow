#!/usr/bin/env bash
# Prépare les secrets Secret Manager pour Stripe (sans clés réelles).
# Usage:
#   ./infra/gcp/setup-stripe-secrets.sh              # crée les secrets vides si absents
#   ./infra/gcp/setup-stripe-secrets.sh --dry-run    # affiche uniquement les instructions
#
# Pour ajouter une vraie valeur (ex. clé live) :
#   echo -n 'sk_live_...' | gcloud secrets versions add petsfollow-stripe-secret-key \
#     --data-file=- --project="${GCP_PROJECT_ID}"
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"

# Secret Manager name → Cloud Run env var (via --set-secrets)
declare -A STRIPE_SECRETS=(
  [petsfollow-stripe-secret-key]=STRIPE_SECRET_KEY
  [petsfollow-stripe-webhook-secret]=STRIPE_WEBHOOK_SECRET
  [petsfollow-stripe-price-annual-onetime]=STRIPE_PRICE_ANNUAL_ONETIME
  [petsfollow-stripe-price-triennial-onetime]=STRIPE_PRICE_TRIENNIAL_ONETIME
  [petsfollow-stripe-price-quinquennial-onetime]=STRIPE_PRICE_QUINQUENNIAL_ONETIME
  [petsfollow-stripe-price-monthly-sub]=STRIPE_PRICE_MONTHLY_SUB
  [petsfollow-stripe-price-annual-sub]=STRIPE_PRICE_ANNUAL_SUB
  [petsfollow-stripe-price-triennial-sub]=STRIPE_PRICE_TRIENNIAL_SUB
)

echo "=== petsFollow Stripe secrets — ${GCP_PROJECT_ID} ==="
echo ""
echo "Secrets requis (Secret Manager → variable Cloud Run) :"
for secret in "${!STRIPE_SECRETS[@]}"; do
  printf '  %-45s → %s\n' "$secret" "${STRIPE_SECRETS[$secret]}"
done
echo ""
echo "Variables d'environnement non secrètes (env-vars-file) :"
echo "  STRIPE_SUCCESS_URL=petsfollow://payment/success"
echo "  STRIPE_CANCEL_URL=petsfollow://payment/cancel"
echo "  BILLING_MOCK_ENABLED=false   # après configuration Stripe live"
echo ""
echo "Exemple --set-secrets Cloud Run (à fusionner avec pf_api_secrets) :"
parts=()
for secret in "${!STRIPE_SECRETS[@]}"; do
  parts+=("${STRIPE_SECRETS[$secret]}=${secret}:latest")
done
IFS=,
echo "  ${parts[*]}"
echo ""

ensure_secret() {
  local name="$1"
  if gcloud secrets describe "$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "  ✓ ${name} existe"
    return 0
  fi
  if [[ "$DRY_RUN" == true ]]; then
    echo "  → CREATE ${name} (dry-run)"
    return 0
  fi
  echo "→ CREATE secret ${name} (placeholder vide)"
  gcloud secrets create "$name" --replication-policy=automatic --project="$GCP_PROJECT_ID" --quiet
  printf '%s' "REPLACE_ME" | gcloud secrets versions add "$name" --data-file=- --project="$GCP_PROJECT_ID" --quiet
}

grant_accessor() {
  local name="$1"
  if [[ "$DRY_RUN" == true ]]; then
    return 0
  fi
  if gcloud secrets describe "$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    gcloud secrets add-iam-policy-binding "$name" \
      --project="$GCP_PROJECT_ID" \
      --member="serviceAccount:${SA_EMAIL}" \
      --role=roles/secretmanager.secretAccessor \
      --quiet >/dev/null 2>&1 || true
  fi
}

echo "→ Vérification / création des secrets"
for secret in "${!STRIPE_SECRETS[@]}"; do
  ensure_secret "$secret"
  grant_accessor "$secret"
done

cat <<EOF

Prochaines étapes (Dashboard Stripe → mode Live) :
  1. Créer les Prices : monthly SUB (3,50 €/mois) · annual ONETIME+SUB · triennial ONETIME+SUB
  2. Copier les price_… IDs dans les secrets petsfollow-stripe-price-*
  3. Webhook endpoint : \${PUBLIC_API_URL}/api/v1/billing/webhook
     Événements : checkout.session.completed, invoice.paid, invoice.payment_failed,
                  customer.subscription.updated, customer.subscription.deleted
  4. Copier whsec_… dans petsfollow-stripe-webhook-secret
  5. Copier sk_live_… dans petsfollow-stripe-secret-key
  6. Customer Portal : activer dans Stripe Dashboard (Settings → Billing → Customer portal)
  7. Redéployer avec BILLING_MOCK_ENABLED=false

Ce script n'écrit jamais de clés réelles — uniquement des placeholders REPLACE_ME.
EOF
