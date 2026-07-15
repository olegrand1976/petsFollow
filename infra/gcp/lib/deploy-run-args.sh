# Variables Cloud Run partagées petsFollow. Source : infra/gcp/lib/gcp-env.sh
# shellcheck shell=bash
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
  local redis_addr
  redis_addr="$(pf_resolve_redis_addr)"
  cat >"$path" <<EOF
HTTP_ADDR: ":8080"
LOG_LEVEL: "info"
MIGRATE_ON_BOOT: "false"
DEV_SEED_ENABLED: "${seed_enabled}"
REDIS_ADDR: "${redis_addr}"
REDIS_KEY_PREFIX: "${REDIS_KEY_PREFIX}:"
SMTP_HOST: "pro1.mail.ovh.net"
SMTP_PORT: "587"
SMTP_FROM: "petsFollow <noreply@ll-it-sc.be>"
PETSFOLLOW_PUBLIC_SITE_URL: "${PUBLIC_SITE_URL}"
PETSFOLLOW_API_PUBLIC_URL: "${PUBLIC_API_URL}"
BILLING_MOCK_ENABLED: "true"
GCS_MEDIA_BUCKET: "${GCS_MEDIA_BUCKET}"
EOF
}

pf_write_frontend_env_file() {
  local path="$1"
  local api_url="${2:-${PUBLIC_API_URL}}"
  cat >"$path" <<EOF
NUXT_PUBLIC_API_BASE: "${api_url}"
NUXT_API_BASE: "${api_url}"
NUXT_PUBLIC_SITE_URL: "${PUBLIC_SITE_URL}"
HOST: "0.0.0.0"
PORT: "3000"
NITRO_PORT: "3000"
EOF
}

pf_api_secrets() {
  printf '%s' "DATABASE_URL=petsfollow-database-url:latest,JWT_SIGNING_KEY=petsfollow-jwt-signing-key:latest"
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
