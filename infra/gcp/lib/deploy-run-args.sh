# Variables Cloud Run partagées petsFollow. Source : infra/gcp/lib/gcp-env.sh
# shellcheck shell=bash
# BILLING_MOCK_ENABLED : défaut true (sécurité). Pour Stripe live :
#   BILLING_MOCK_ENABLED=false ./infra/gcp/cloudbuild.yaml …
# ou export avant deploy manuel. Voir documentation/07-STRIPE-BILLING.md
set -euo pipefail

pf_resolve_redis_addr() {
  local redis_url host port
  redis_url="$(gcloud secrets versions access latest \
    --secret=petsfollow-redis-url --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -n "$redis_url" ]]; then
    host="$(python3 -c "from urllib.parse import urlparse; u=urlparse('$redis_url'); print(u.hostname or '')")"
    port="$(python3 -c "from urllib.parse import urlparse; u=urlparse('$redis_url'); print(u.port or 6379)")"
    if [[ -n "$host" ]]; then
      printf '%s:%s' "$host" "$port"
      return 0
    fi
  fi
  local vm_host
  vm_host="$(gcloud secrets versions access latest \
    --secret=premedica-redis-host --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$vm_host" ]]; then
    vm_host="$(gcloud compute instances describe "$REDIS_VM_NAME" \
      --zone="$REDIS_VM_ZONE" --project="$GCP_PROJECT_ID" \
      --format='value(networkInterfaces[0].networkIP)' 2>/dev/null || true)"
  fi
  printf '%s:6379' "${vm_host:-10.200.0.2}"
}

pf_write_api_env_file() {
  local path="$1"
  local seed_enabled="${2:-false}"
  local admin_staging_seed="${3:-false}"
  # Explicit : staging | production — défaut production (seed/seed-mass refusés
  # hors allowlist, donc un déploiement sans APP_ENV ne peut pas tronquer la base).
  local app_env="${4:-production}"
  local redis_addr
  local billing_mock
  local pharmacy_enabled billit_enabled prescriptions_enabled
  billing_mock="${BILLING_MOCK_ENABLED:-true}"
  redis_addr="$(pf_resolve_redis_addr)"
  # Modules tag « dev » : on en staging (sidebar Pro) ; prod reste opt-in explicite.
  local billit_mock="false"
  local billit_secrets_backend=""
  if [[ "$app_env" == "staging" ]]; then
    pharmacy_enabled="${PHARMACY_ENABLED:-true}"
    billit_enabled="${BILLIT_ENABLED:-true}"
    prescriptions_enabled="${PRESCRIPTIONS_ENABLED:-true}"
    # Billit tag-dev : mock gateway (pas d’appels live). local_enc + clé via pf_api_secrets.
    if [[ "$billit_enabled" == "true" || "$billit_enabled" == "1" ]]; then
      billit_mock="${BILLIT_MOCK_ENABLED:-true}"
      billit_secrets_backend="${BILLIT_SECRETS_BACKEND:-local_enc}"
    fi
  else
    pharmacy_enabled="${PHARMACY_ENABLED:-false}"
    billit_enabled="${BILLIT_ENABLED:-false}"
    prescriptions_enabled="${PRESCRIPTIONS_ENABLED:-false}"
    billit_mock="${BILLIT_MOCK_ENABLED:-false}"
    if [[ "$billit_enabled" == "true" || "$billit_enabled" == "1" ]]; then
      billit_secrets_backend="${BILLIT_SECRETS_BACKEND:-local_enc}"
    fi
  fi
  cat >"$path" <<EOF
HTTP_ADDR: ":8080"
LOG_LEVEL: "info"
APP_ENV: "${app_env}"
MIGRATE_ON_BOOT: "false"
DEV_SEED_ENABLED: "${seed_enabled}"
ADMIN_STAGING_SEED_ENABLED: "${admin_staging_seed}"
REDIS_ADDR: "${redis_addr}"
REDIS_KEY_PREFIX: "${REDIS_KEY_PREFIX}:"
SMTP_HOST: "pro1.mail.ovh.net"
SMTP_PORT: "587"
SMTP_FROM: "petsFollow <noreply@petsfollow.app>"
SMTP_USER: "noreply@petsfollow.app"
OPS_NOTIFY_EMAIL: "${OPS_NOTIFY_EMAIL:-o.legrand1976@gmail.com}"
SUPPORT_INBOX_EMAIL: "${SUPPORT_INBOX_EMAIL:-o.legrand1976@gmail.com}"
# Fallback téléphone commercial (mail/PDF envoi dossier) si profil commercial vide.
COMMERCIAL_CONTACT_PHONE: "${COMMERCIAL_CONTACT_PHONE:-}"
PETSFOLLOW_PUBLIC_SITE_URL: "${PUBLIC_SITE_URL}"
PETSFOLLOW_API_PUBLIC_URL: "${PUBLIC_API_URL}"
BILLING_MOCK_ENABLED: "${billing_mock}"
PHARMACY_ENABLED: "${pharmacy_enabled}"
BILLIT_ENABLED: "${billit_enabled}"
BILLIT_MOCK_ENABLED: "${billit_mock}"
PRESCRIPTIONS_ENABLED: "${prescriptions_enabled}"
GCS_MEDIA_BUCKET: "${GCS_MEDIA_BUCKET}"
LLIT_WEBSITE_URL: "${LLIT_WEBSITE_URL:-https://ll-it-sc.be}"
GEMINI_MODEL: "${GEMINI_MODEL:-gemini-3.6-flash}"
GEMINI_LITE_MODEL: "${GEMINI_LITE_MODEL:-gemini-3.5-flash-lite}"
GEMINI_LIVE_MODEL: "${GEMINI_LIVE_MODEL:-gemini-2.5-flash-native-audio-preview-09-2025}"
GOOGLE_OAUTH_CLIENT_ID: "${GOOGLE_OAUTH_CLIENT_ID:-237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com}"
EOF
  if [[ -n "$billit_secrets_backend" ]]; then
    cat >>"$path" <<EOF
BILLIT_SECRETS_BACKEND: "${billit_secrets_backend}"
EOF
  fi
}

pf_write_frontend_env_file() {
  local path="$1"
  local api_url="${2:-${PUBLIC_API_URL}}"
  # Explicit : staging | production | local — défaut production (jamais activer UC/badge S par accident).
  local app_env="${3:-production}"
  local pharmacy_pub billit_pub prescriptions_pub
  local flag_lines=""
  # Nav Pro tag « dev » : on en staging ; prod opt-in.
  # Ne jamais écrire "false" : Nuxt injecte des strings et Boolean("false")===true côté JS.
  # Clé absente → boolean bake (false en image prod).
  if [[ "$app_env" == "staging" ]]; then
    pharmacy_pub="${NUXT_PUBLIC_PHARMACY_ENABLED:-true}"
    billit_pub="${NUXT_PUBLIC_BILLIT_ENABLED:-true}"
    prescriptions_pub="${NUXT_PUBLIC_PRESCRIPTIONS_ENABLED:-true}"
  else
    pharmacy_pub="${NUXT_PUBLIC_PHARMACY_ENABLED:-}"
    billit_pub="${NUXT_PUBLIC_BILLIT_ENABLED:-}"
    prescriptions_pub="${NUXT_PUBLIC_PRESCRIPTIONS_ENABLED:-}"
  fi
  if [[ "$pharmacy_pub" == "true" || "$pharmacy_pub" == "1" ]]; then
    flag_lines="${flag_lines}NUXT_PUBLIC_PHARMACY_ENABLED: \"true\"
"
  fi
  if [[ "$billit_pub" == "true" || "$billit_pub" == "1" ]]; then
    flag_lines="${flag_lines}NUXT_PUBLIC_BILLIT_ENABLED: \"true\"
"
  fi
  if [[ "$prescriptions_pub" == "true" || "$prescriptions_pub" == "1" ]]; then
    flag_lines="${flag_lines}NUXT_PUBLIC_PRESCRIPTIONS_ENABLED: \"true\"
"
  fi
  cat >"$path" <<EOF
NUXT_PUBLIC_API_BASE: "${api_url}"
NUXT_API_BASE: "${api_url}"
NUXT_PUBLIC_SITE_URL: "${PUBLIC_SITE_URL}"
NUXT_PUBLIC_GOOGLE_CLIENT_ID: "${GOOGLE_OAUTH_CLIENT_ID:-237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com}"
NUXT_PUBLIC_APP_ENV: "${app_env}"
${flag_lines}HOST: "0.0.0.0"
NITRO_PORT: "3000"
NODE_OPTIONS: "--max-old-space-size=768"
EOF
}

pf_api_secrets() {
  local secrets="DATABASE_URL=petsfollow-database-url:latest,JWT_SIGNING_KEY=petsfollow-jwt-signing-key:latest"
  if gcloud secrets versions access latest \
    --secret=petsfollow-smtp-password --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},SMTP_PASS=petsfollow-smtp-password:latest"
  fi
  if gcloud secrets versions access latest \
    --secret=petsfollow-auth-health-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},AUTH_HEALTH_SECRET=petsfollow-auth-health-secret:latest"
  fi
  # Import clients admin — mapping colonnes (Secret Manager).
  if gcloud secrets versions access latest \
    --secret=petsfollow-gemini-api-key --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},GEMINI_API_KEY=petsfollow-gemini-api-key:latest"
  fi
  if gcloud secrets versions access latest \
    --secret=petsfollow-product-digest-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},PRODUCT_DIGEST_SECRET=petsfollow-product-digest-secret:latest"
  fi
  if gcloud secrets versions access latest \
    --secret=petsfollow-pitch-analyzer-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},PITCH_ANALYZER_SECRET=petsfollow-pitch-analyzer-secret:latest"
  fi
  if gcloud secrets versions access latest \
    --secret=petsfollow-ai-module-friction-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},AI_MODULE_FRICTION_SECRET=petsfollow-ai-module-friction-secret:latest"
  fi
  # Sans ce secret, /internal/retention/run répond 401 : la purge RGPD ne tourne pas.
  if gcloud secrets versions access latest \
    --secret=petsfollow-retention-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},RETENTION_PURGE_SECRET=petsfollow-retention-secret:latest"
  fi
  if gcloud secrets versions access latest \
    --secret=petsfollow-sales-branches-auto-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},SALES_BRANCHES_AUTO_SECRET=petsfollow-sales-branches-auto-secret:latest"
  fi
  if gcloud secrets versions access latest \
    --secret=petsfollow-saas-invoices-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},SAAS_INVOICES_SECRET=petsfollow-saas-invoices-secret:latest"
  fi
  # Sans ce secret, /internal/pharmacy/expiry-run répond 401 : pas d'auto-quarantaine.
  if gcloud secrets versions access latest \
    --secret=petsfollow-pharmacy-expiry-secret --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},PHARMACY_EXPIRY_SECRET=petsfollow-pharmacy-expiry-secret:latest"
  fi
  # Billit local_enc (staging tag-dev mock / live) — clé dédiée si présente, sinon JWT.
  if gcloud secrets versions access latest \
    --secret=petsfollow-billit-secrets-key --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},BILLIT_SECRETS_KEY=petsfollow-billit-secrets-key:latest"
  else
    secrets="${secrets},BILLIT_SECRETS_KEY=petsfollow-jwt-signing-key:latest"
  fi
  printf '%s' "$secrets"
}

pf_migrate_secrets() {
  if gcloud secrets versions access latest \
    --secret=petsfollow-migrate-database-url --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    printf '%s' "DATABASE_URL=petsfollow-migrate-database-url:latest,JWT_SIGNING_KEY=petsfollow-jwt-signing-key:latest"
  else
    echo "→ Job migrate : fallback petsfollow-database-url" >&2
    printf '%s' "DATABASE_URL=petsfollow-database-url:latest,JWT_SIGNING_KEY=petsfollow-jwt-signing-key:latest"
  fi
}
