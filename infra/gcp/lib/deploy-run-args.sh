# Variables Cloud Run partagées petsFollow. Source : infra/gcp/lib/gcp-env.sh
# shellcheck shell=bash
# BILLING_MOCK_ENABLED : défaut true (sécurité). Pour Stripe live :
#   BILLING_MOCK_ENABLED=false ./infra/gcp/cloudbuild.yaml …
# ou export avant deploy manuel. Voir documentation/07-STRIPE-BILLING.md
set -euo pipefail

pf_sm_database_url() { printf '%s' "${PF_SM_DATABASE_URL:-petsfollow-database-url}"; }
pf_sm_migrate_database_url() { printf '%s' "${PF_SM_MIGRATE_DATABASE_URL:-petsfollow-migrate-database-url}"; }
pf_sm_jwt_signing_key() { printf '%s' "${PF_SM_JWT_SIGNING_KEY:-petsfollow-jwt-signing-key}"; }
pf_sm_redis_url() { printf '%s' "${PF_SM_REDIS_URL:-petsfollow-redis-url}"; }
pf_sm_billit_secrets_key() { printf '%s' "${PF_SM_BILLIT_SECRETS_KEY:-petsfollow-billit-secrets-key}"; }
pf_sm_billit_webhook_secret() { printf '%s' "${PF_SM_BILLIT_WEBHOOK_SECRET:-petsfollow-billit-webhook-secret}"; }
pf_sm_billit_master_api_key() { printf '%s' "${PF_SM_BILLIT_MASTER_API_KEY:-petsfollow-billit-master-api-key}"; }

# Job secrets : staging petsfollow-<name> ; prod petsfollow-prod-<name>
pf_sm_job_secret() {
  local name="$1"
  if [[ "${PETSFOLLOW_GCP_ENV:-}" == "prod" ]]; then
    printf 'petsfollow-prod-%s' "$name"
  else
    printf 'petsfollow-%s' "$name"
  fi
}

pf_sm_has_secret() {
  local name="$1"
  gcloud secrets versions access latest --secret="$name" --project="$GCP_PROJECT_ID" >/dev/null 2>&1
}

pf_api_mount_job_secret() {
  # $1 = env var name (PRODUCT_DIGEST_SECRET), $2 = SM short name (product-digest-secret)
  local env_name="$1"
  local short="$2"
  local sm
  sm="$(pf_sm_job_secret "$short")"
  if pf_sm_has_secret "$sm"; then
    printf ',%s=%s:latest' "$env_name" "$sm"
  fi
}

pf_resolve_redis_addr() {
  local redis_url host port redis_secret
  redis_secret="$(pf_sm_redis_url)"
  redis_url="$(gcloud secrets versions access latest \
    --secret="$redis_secret" --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  if [[ -z "$redis_url" && "$redis_secret" != "petsfollow-redis-url" ]]; then
    redis_url="$(gcloud secrets versions access latest \
      --secret=petsfollow-redis-url --project="$GCP_PROJECT_ID" 2>/dev/null || true)"
  fi
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
  local pharmacy_enabled billit_enabled prescriptions_enabled pacs_enabled research_enabled vet_news_enabled client_ai_enabled
  local ai_cr_advanced_enabled sms_enabled crewai_base_url crewai_use_id_token
  billing_mock="${BILLING_MOCK_ENABLED:-true}"
  redis_addr="$(pf_resolve_redis_addr)"
  # Modules tag « dev » : on en staging (sidebar Pro) ; prod reste opt-in explicite.
  # Billit Access Point : staging → sandbox API ; prod (main) → api.billit.be (whitelist).
  # Docs : https://docs.accesspoint.billit.eu/docs/sandbox-vs-production
  local billit_mock="false"
  local billit_secrets_backend=""
  local billit_base_url=""
  local billit_reseller_url=""
  local billit_webhook_sm
  billit_webhook_sm="$(pf_sm_billit_webhook_secret)"
  if [[ "$app_env" == "staging" ]]; then
    pharmacy_enabled="${PHARMACY_ENABLED:-true}"
    billit_enabled="${BILLIT_ENABLED:-true}"
    prescriptions_enabled="${PRESCRIPTIONS_ENABLED:-true}"
    research_enabled="${RESEARCH_ENABLED:-true}"
    vet_news_enabled="${VET_NEWS_ENABLED:-true}"
    client_ai_enabled="${CLIENT_AI_ENABLED:-true}"
    ai_cr_advanced_enabled="${AI_CR_ADVANCED_ENABLED:-true}"
    sms_enabled="${SMS_ENABLED:-true}"
    crewai_base_url="${CREWAI_BASE_URL:-${CREWAI_STAGING_URL:-https://crewai-orchestrator-staging-a7ako2njea-od.a.run.app}}"
    crewai_use_id_token="${CREWAI_USE_ID_TOKEN:-true}"
    # PACS on staging only when Orthanc URL is wired (avoid permanent offline UI).
    if [[ -n "${PACS_ORTHANC_URL:-}" ]]; then
      pacs_enabled="${PACS_ENABLED:-true}"
    else
      pacs_enabled="${PACS_ENABLED:-false}"
    fi
    # Billit staging = sandbox. Live mock-off only when webhook secret is in SM
    # (ValidateBillit refuses live without BILLIT_WEBHOOK_SECRET) — sinon mock.
    if [[ "$billit_enabled" == "true" || "$billit_enabled" == "1" ]]; then
      billit_base_url="${BILLIT_BASE_URL:-https://api.sandbox.billit.be}"
      billit_reseller_url="${BILLIT_RESELLER_REGISTER_URL:-https://my.sandbox.billit.be/Account/Register}"
      billit_secrets_backend="${BILLIT_SECRETS_BACKEND:-local_enc}"
      if [[ -n "${BILLIT_MOCK_ENABLED:-}" ]]; then
        billit_mock="${BILLIT_MOCK_ENABLED}"
      elif pf_sm_has_secret "$billit_webhook_sm"; then
        billit_mock="false"
      else
        billit_mock="true"
      fi
    fi
  else
    pharmacy_enabled="${PHARMACY_ENABLED:-false}"
    billit_enabled="${BILLIT_ENABLED:-false}"
    prescriptions_enabled="${PRESCRIPTIONS_ENABLED:-false}"
    pacs_enabled="${PACS_ENABLED:-false}"
    research_enabled="${RESEARCH_ENABLED:-false}"
    vet_news_enabled="${VET_NEWS_ENABLED:-false}"
    client_ai_enabled="${CLIENT_AI_ENABLED:-false}"
    ai_cr_advanced_enabled="${AI_CR_ADVANCED_ENABLED:-false}"
    sms_enabled="${SMS_ENABLED:-false}"
    crewai_base_url="${CREWAI_BASE_URL:-}"
    crewai_use_id_token="${CREWAI_USE_ID_TOKEN:-false}"
    billit_mock="${BILLIT_MOCK_ENABLED:-false}"
    if [[ "$billit_enabled" == "true" || "$billit_enabled" == "1" ]]; then
      billit_base_url="${BILLIT_BASE_URL:-https://api.billit.be}"
      billit_reseller_url="${BILLIT_RESELLER_REGISTER_URL:-https://my.billit.be/account/PetsFollow/Register}"
      billit_secrets_backend="${BILLIT_SECRETS_BACKEND:-local_enc}"
    fi
  fi
  cat >"$path" <<EOF
HTTP_ADDR: ":8080"
LOG_LEVEL: "info"
APP_ENV: "${app_env}"
# GC : sans limite connue, le runtime laisse le tas grossir jusqu'à l'OOM kill de
# Cloud Run (--memory=1Gi). 800MiB garde ~200 Mo hors tas (stacks, buffers, tzdata)
# et laisse le GC Green Tea de Go 1.26 arbitrer avant la mort du conteneur.
GOMEMLIMIT: "${GOMEMLIMIT:-800MiB}"
# Un seul hop (l'orchestrateur Cloud Run) devant le conteneur : seule la dernière
# entrée de X-Forwarded-For n'est pas fournie par l'appelant. Voir httpx.ClientIP.
# Trafic Pro (BFF) : IP client via X-PF-Client-IP + BFF_PROXY_SECRET (SM optionnel).
TRUSTED_PROXY_HOPS: "${TRUSTED_PROXY_HOPS:-1}"
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
SUPPORT_INBOX_EMAIL: "${SUPPORT_INBOX_EMAIL:-barbara@petsfollow.app}"
# Fallback téléphone commercial (mail/PDF envoi dossier) si profil commercial vide.
COMMERCIAL_CONTACT_PHONE: "${COMMERCIAL_CONTACT_PHONE:-0478.02.33.77}"
PETSFOLLOW_PUBLIC_SITE_URL: "${PUBLIC_SITE_URL}"
PETSFOLLOW_API_PUBLIC_URL: "${PUBLIC_API_URL}"
CORS_ALLOWED_ORIGINS: "${CORS_ALLOWED_ORIGINS:-${PUBLIC_SITE_URL}}"
BILLING_MOCK_ENABLED: "${billing_mock}"
PHARMACY_ENABLED: "${pharmacy_enabled}"
# Déclarations DAF : dry-run FORCÉ (staging + prod). Une clé software-house (listes)
# ne permet pas le live declare — voir documentation/39-VAMREG-AFMPS-READONLY.md.
# Live declare = ICD write + credentials déclarant + retirer ce forçage volontairement.
VAMREG_DRY_RUN: "true"
PHARMACY_WORKERS_ENABLED: "${PHARMACY_WORKERS_ENABLED:-false}"
BILLIT_ENABLED: "${billit_enabled}"
BILLIT_MOCK_ENABLED: "${billit_mock}"
PRESCRIPTIONS_ENABLED: "${prescriptions_enabled}"
PACS_ENABLED: "${pacs_enabled}"
RESEARCH_ENABLED: "${research_enabled}"
VET_NEWS_ENABLED: "${vet_news_enabled}"
CLIENT_AI_ENABLED: "${client_ai_enabled}"
AI_CR_ADVANCED_ENABLED: "${ai_cr_advanced_enabled}"
SMS_ENABLED: "${sms_enabled}"
# Envoi SMS : dry-run FORCÉ (staging + prod) tant que TELNYX_API_KEY / la clé
# publique webhook ne sont pas montées en Secret Manager. Passer à false est un
# geste volontaire : cela déclenche des envois facturés et exige TELNYX_PUBLIC_KEY
# (sinon un STOP entrant serait perdu) — voir documentation/44-SMS-TELNYX.md.
SMS_DRY_RUN: "${SMS_DRY_RUN:-true}"
SMS_DEFAULT_REGION: "${SMS_DEFAULT_REGION:-BE}"
TELNYX_MESSAGING_PROFILE_ID: "${TELNYX_MESSAGING_PROFILE_ID:-}"
TELNYX_FROM: "${TELNYX_FROM:-}"
PACS_ORTHANC_URL: "${PACS_ORTHANC_URL:-}"
PACS_ORTHANC_USER: "${PACS_ORTHANC_USER:-petsfollow}"
PACS_ORTHANC_USE_ID_TOKEN: "${PACS_ORTHANC_USE_ID_TOKEN:-true}"
GCS_MEDIA_BUCKET: "${GCS_MEDIA_BUCKET}"
LLIT_WEBSITE_URL: "${LLIT_WEBSITE_URL:-https://ll-it-sc.be}"
GEMINI_MODEL: "${GEMINI_MODEL:-gemini-3.6-flash}"
GEMINI_LITE_MODEL: "${GEMINI_LITE_MODEL:-gemini-3.5-flash-lite}"
GEMINI_LIVE_MODEL: "${GEMINI_LIVE_MODEL:-gemini-2.5-flash-native-audio-preview-09-2025}"
GEMINI_EMBEDDING_MODEL: "${GEMINI_EMBEDDING_MODEL:-text-embedding-004}"
CREWAI_BASE_URL: "${crewai_base_url}"
CREWAI_USE_ID_TOKEN: "${crewai_use_id_token}"
GOOGLE_OAUTH_CLIENT_ID: "${GOOGLE_OAUTH_CLIENT_ID:-237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com}"
EOF
  if [[ -n "${VAMREG_BASE_URL:-}" ]]; then
    cat >>"$path" <<EOF
VAMREG_BASE_URL: "${VAMREG_BASE_URL}"
EOF
  fi
  # Readonly AFMPS reference lists (ICD software-house) — séparée de VAMREG_BASE_URL (déclaration).
  cat >>"$path" <<EOF
VAMREG_AFMPS_BASE_URL: "${VAMREG_AFMPS_BASE_URL:-https://app.fagg-afmps.be/vamreg/api}"
EOF
  if [[ -n "$billit_secrets_backend" ]]; then
    cat >>"$path" <<EOF
BILLIT_SECRETS_BACKEND: "${billit_secrets_backend}"
EOF
  fi
  if [[ -n "$billit_base_url" ]]; then
    cat >>"$path" <<EOF
BILLIT_BASE_URL: "${billit_base_url}"
EOF
  fi
  if [[ -n "$billit_reseller_url" ]]; then
    cat >>"$path" <<EOF
BILLIT_RESELLER_REGISTER_URL: "${billit_reseller_url}"
EOF
  fi
  # PartyID master (non secret) — Flux A SaaS + header Access Point en prod.
  if [[ -n "${BILLIT_MASTER_PARTY_ID:-}" ]]; then
    cat >>"$path" <<EOF
BILLIT_MASTER_PARTY_ID: "${BILLIT_MASTER_PARTY_ID}"
EOF
  fi
}

pf_write_frontend_env_file() {
  local path="$1"
  local api_url="${2:-${PUBLIC_API_URL}}"
  # Explicit : staging | production | local — défaut production (jamais activer UC/badge S par accident).
  local app_env="${3:-production}"
  local pharmacy_pub billit_pub prescriptions_pub pacs_pub research_pub vet_news_pub sites_pub ai_cr_pub
  local flag_lines=""
  # Nav Pro tag « dev » : on en staging ; prod opt-in.
  # Ne jamais écrire "false" : Nuxt injecte des strings et Boolean("false")===true côté JS.
  # Clé absente → boolean bake (false en image prod).
  if [[ "$app_env" == "staging" ]]; then
    pharmacy_pub="${NUXT_PUBLIC_PHARMACY_ENABLED:-true}"
    billit_pub="${NUXT_PUBLIC_BILLIT_ENABLED:-true}"
    prescriptions_pub="${NUXT_PUBLIC_PRESCRIPTIONS_ENABLED:-true}"
    research_pub="${NUXT_PUBLIC_RESEARCH_ENABLED:-true}"
    vet_news_pub="${NUXT_PUBLIC_VET_NEWS_ENABLED:-true}"
    sites_pub="${NUXT_PUBLIC_SITES_UI_ENABLED:-true}"
    ai_cr_pub="${NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED:-true}"
    if [[ -n "${PACS_ORTHANC_URL:-}" ]]; then
      pacs_pub="${NUXT_PUBLIC_PACS_ENABLED:-true}"
    else
      pacs_pub="${NUXT_PUBLIC_PACS_ENABLED:-}"
    fi
  else
    pharmacy_pub="${NUXT_PUBLIC_PHARMACY_ENABLED:-}"
    billit_pub="${NUXT_PUBLIC_BILLIT_ENABLED:-}"
    prescriptions_pub="${NUXT_PUBLIC_PRESCRIPTIONS_ENABLED:-}"
    pacs_pub="${NUXT_PUBLIC_PACS_ENABLED:-}"
    research_pub="${NUXT_PUBLIC_RESEARCH_ENABLED:-}"
    vet_news_pub="${NUXT_PUBLIC_VET_NEWS_ENABLED:-}"
    sites_pub="${NUXT_PUBLIC_SITES_UI_ENABLED:-}"
    ai_cr_pub="${NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED:-}"
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
  if [[ "$pacs_pub" == "true" || "$pacs_pub" == "1" ]]; then
    flag_lines="${flag_lines}NUXT_PUBLIC_PACS_ENABLED: \"true\"
"
  fi
  if [[ "$research_pub" == "true" || "$research_pub" == "1" ]]; then
    flag_lines="${flag_lines}NUXT_PUBLIC_RESEARCH_ENABLED: \"true\"
"
  fi
  if [[ "$vet_news_pub" == "true" || "$vet_news_pub" == "1" ]]; then
    flag_lines="${flag_lines}NUXT_PUBLIC_VET_NEWS_ENABLED: \"true\"
"
  fi
  if [[ "$sites_pub" == "true" || "$sites_pub" == "1" ]]; then
    flag_lines="${flag_lines}NUXT_PUBLIC_SITES_UI_ENABLED: \"true\"
"
  fi
  if [[ "$ai_cr_pub" == "true" || "$ai_cr_pub" == "1" ]]; then
    flag_lines="${flag_lines}NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED: \"true\"
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

# Secret BFF→API (X-PF-Client-IP). Absent = Nuxt ne relaie pas l'IP client
# (rate limit Pro tombe sur l'egress Nuxt — dégradé mais non spoofable).
pf_nuxt_secrets() {
  local secrets=""
  secrets="${secrets}$(pf_api_mount_job_secret BFF_PROXY_SECRET bff-proxy-secret)"
  # Trim leading comma if present
  secrets="${secrets#,}"
  printf '%s' "$secrets"
}

pf_api_secrets() {
  local db_secret jwt_secret
  db_secret="$(pf_sm_database_url)"
  jwt_secret="$(pf_sm_jwt_signing_key)"
  local secrets="DATABASE_URL=${db_secret}:latest,JWT_SIGNING_KEY=${jwt_secret}:latest"
  if gcloud secrets versions access latest \
    --secret=petsfollow-smtp-password --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},SMTP_PASS=petsfollow-smtp-password:latest"
  fi
  # Job secrets : préfixe petsfollow- / petsfollow-prod- selon PETSFOLLOW_GCP_ENV.
  secrets="${secrets}$(pf_api_mount_job_secret AUTH_HEALTH_SECRET auth-health-secret)"
  # SMS Telnyx : clé API d'envoi, clé publique de vérification des webhooks,
  # secret du cron rappel J-1. Chacun optionnel — absent = fonctionnalité inactive
  # (pas d'envoi live, webhooks rejetés, cron 401) plutôt qu'un déploiement cassé.
  if gcloud secrets versions access latest \
    --secret=petsfollow-telnyx-api-key --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},TELNYX_API_KEY=petsfollow-telnyx-api-key:latest"
  fi
  if gcloud secrets versions access latest \
    --secret=petsfollow-telnyx-public-key --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},TELNYX_PUBLIC_KEY=petsfollow-telnyx-public-key:latest"
  fi
  secrets="${secrets}$(pf_api_mount_job_secret VISIT_REMINDERS_SECRET visit-reminders-secret)"
  # Import clients admin — mapping colonnes (Secret Manager).
  if gcloud secrets versions access latest \
    --secret=petsfollow-gemini-api-key --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},GEMINI_API_KEY=petsfollow-gemini-api-key:latest"
  fi
  secrets="${secrets}$(pf_api_mount_job_secret PRODUCT_DIGEST_SECRET product-digest-secret)"
  secrets="${secrets}$(pf_api_mount_job_secret PITCH_ANALYZER_SECRET pitch-analyzer-secret)"
  secrets="${secrets}$(pf_api_mount_job_secret AI_MODULE_FRICTION_SECRET ai-module-friction-secret)"
  # Sans ce secret, /internal/retention/run répond 401 : la purge RGPD ne tourne pas.
  secrets="${secrets}$(pf_api_mount_job_secret RETENTION_PURGE_SECRET retention-secret)"
  secrets="${secrets}$(pf_api_mount_job_secret BFF_PROXY_SECRET bff-proxy-secret)"
  secrets="${secrets}$(pf_api_mount_job_secret SALES_BRANCHES_AUTO_SECRET sales-branches-auto-secret)"
  secrets="${secrets}$(pf_api_mount_job_secret SAAS_INVOICES_SECRET saas-invoices-secret)"
  # Sans ce secret, /internal/invoicing-reconcile/run répond 401 : une facture
  # dont le webhook s'est perdu reste en vol jusqu'au rejet automatique à 7 jours.
  secrets="${secrets}$(pf_api_mount_job_secret INVOICING_RECONCILE_SECRET invoicing-reconcile-secret)"
  # Sans ce secret, /internal/pharmacy/expiry-run répond 401 : pas d'auto-quarantaine.
  secrets="${secrets}$(pf_api_mount_job_secret PHARMACY_EXPIRY_SECRET pharmacy-expiry-secret)"
  # Research ETL + salt HMAC (observatoire) — requis si RESEARCH_ENABLED hors seedable.
  secrets="${secrets}$(pf_api_mount_job_secret RESEARCH_ETL_SECRET research-etl-secret)"
  secrets="${secrets}$(pf_api_mount_job_secret RESEARCH_ANON_SALT research-anon-salt)"
  secrets="${secrets}$(pf_api_mount_job_secret VET_NEWS_SECRET vet-news-secret)"
  # Sans ce secret, /internal/rag/reindex et /internal/rag/search répondent 401.
  secrets="${secrets}$(pf_api_mount_job_secret RAG_REINDEX_SECRET rag-reindex-secret)"
  # Shared CrewAI orchestrator (X-Crew-Secret). Prefer petsfollow-prefixed job secret;
  # fallback legacy CREWAI_WEBHOOK_SECRET_STAGING is mounted via CREWAI_SHARED_SECRET_SM override.
  secrets="${secrets}$(pf_api_mount_job_secret CREWAI_SHARED_SECRET crewai-shared-secret)"
  if [[ -n "${CREWAI_SHARED_SECRET_SM:-}" ]] && pf_sm_has_secret "${CREWAI_SHARED_SECRET_SM}"; then
    secrets="${secrets},CREWAI_SHARED_SECRET=${CREWAI_SHARED_SECRET_SM}:latest"
  elif [[ "${PETSFOLLOW_GCP_ENV:-}" != "prod" ]] && gcloud secrets versions access latest \
    --secret=CREWAI_WEBHOOK_SECRET_STAGING --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    # Staging shared secret used by the platform orchestrator (multi-app).
    secrets="${secrets},CREWAI_SHARED_SECRET=CREWAI_WEBHOOK_SECRET_STAGING:latest"
  fi
  # VAMReg déclaration live (P0-1 write) — NE PAS monter tant que VAMREG_DRY_RUN forcé true.
  # La clé software-house va dans petsfollow-vamreg-afmps-api-key (listes), pas ici.
  if gcloud secrets versions access latest \
    --secret=petsfollow-vamreg-api-key --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    echo "→ Note: petsfollow-vamreg-api-key présent mais non monté (déclarations dry-run)" >&2
  fi
  # VAMReg AFMPS readonly (ICD software-house, FAMHP-SEC-KEY) — staging et prod.
  if gcloud secrets versions access latest \
    --secret=petsfollow-vamreg-afmps-api-key --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},VAMREG_AFMPS_API_KEY=petsfollow-vamreg-afmps-api-key:latest"
  fi
  # Billit local_enc + live sandbox/prod — secrets optionnels (absents = mock ou module off).
  # Staging : petsfollow-billit-* ; prod : PF_SM_BILLIT_* → petsfollow-prod-billit-* (gcp-env-prod).
  local billit_secrets_key_sm billit_webhook_sm billit_master_sm
  billit_secrets_key_sm="$(pf_sm_billit_secrets_key)"
  billit_webhook_sm="$(pf_sm_billit_webhook_secret)"
  billit_master_sm="$(pf_sm_billit_master_api_key)"
  if pf_sm_has_secret "$billit_secrets_key_sm"; then
    secrets="${secrets},BILLIT_SECRETS_KEY=${billit_secrets_key_sm}:latest"
  else
    secrets="${secrets},BILLIT_SECRETS_KEY=${jwt_secret}:latest"
  fi
  if pf_sm_has_secret "$billit_webhook_sm"; then
    secrets="${secrets},BILLIT_WEBHOOK_SECRET=${billit_webhook_sm}:latest"
  fi
  if pf_sm_has_secret "$billit_master_sm"; then
    secrets="${secrets},BILLIT_MASTER_API_KEY=${billit_master_sm}:latest"
  fi
  if gcloud secrets versions access latest \
    --secret=petsfollow-orthanc-password --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    secrets="${secrets},PACS_ORTHANC_PASSWORD=petsfollow-orthanc-password:latest"
  fi
  printf '%s' "$secrets"
}

pf_migrate_secrets() {
  local migrate_secret db_secret jwt_secret
  migrate_secret="$(pf_sm_migrate_database_url)"
  db_secret="$(pf_sm_database_url)"
  jwt_secret="$(pf_sm_jwt_signing_key)"
  if gcloud secrets versions access latest \
    --secret="$migrate_secret" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
    printf '%s' "DATABASE_URL=${migrate_secret}:latest,JWT_SIGNING_KEY=${jwt_secret}:latest"
  else
    echo "→ Job migrate : fallback ${db_secret}" >&2
    printf '%s' "DATABASE_URL=${db_secret}:latest,JWT_SIGNING_KEY=${jwt_secret}:latest"
  fi
}
