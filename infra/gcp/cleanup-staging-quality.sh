#!/usr/bin/env bash
# Lance cleanup-staging-quality contre Cloud SQL staging (Secret Manager + Auth Proxy).
# Usage (local ops ou CI après Playwright staging) :
#   bash infra/gcp/cleanup-staging-quality.sh
#   bash infra/gcp/cleanup-staging-quality.sh --dry-run
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CLEANUP_ARGS=()
for arg in "$@"; do
  case "$arg" in
    --dry-run|--skip-invoicing) CLEANUP_ARGS+=("$arg") ;;
    -h|--help)
      sed -n '2,6p' "$0"
      exit 0
      ;;
  esac
done

if [[ "$CLOUDSQL_INSTANCE" == *"petsfollow-db-prod"* ]] || [[ "$SQL_INSTANCE" == "petsfollow-db-prod" ]]; then
  echo "Refusing quality cleanup against prod Cloud SQL ($CLOUDSQL_INSTANCE)." >&2
  exit 1
fi
if [[ "$CLOUDSQL_INSTANCE" != *"premedica-db-staging"* ]]; then
  echo "Refusing Cloud SQL instance (expected premedica-db-staging): $CLOUDSQL_INSTANCE" >&2
  exit 1
fi

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

PROXY_BIN="${TMPDIR:-/tmp}/cloud-sql-proxy-pf"
PROXY_PORT="${PF_CLEANUP_PROXY_PORT:-55432}"
PROXY_URL="https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.15.2/cloud-sql-proxy.linux.amd64"

if [[ ! -x "$PROXY_BIN" ]]; then
  echo "→ Download cloud-sql-proxy"
  curl -fsSL -o "$PROXY_BIN" "$PROXY_URL"
  chmod +x "$PROXY_BIN"
fi

echo "→ Fetch migrate DATABASE_URL (Secret Manager)"
RAW_URL="$(gcloud secrets versions access latest \
  --secret=petsfollow-migrate-database-url \
  --project="$GCP_PROJECT_ID")"

if [[ "$RAW_URL" == *"petsfollow-db-prod"* ]]; then
  echo "Secret petsfollow-migrate-database-url points at prod — abort." >&2
  exit 1
fi
if [[ "$RAW_URL" != *"premedica-db-staging"* ]]; then
  echo "Secret petsfollow-migrate-database-url missing premedica-db-staging — abort." >&2
  exit 1
fi

# Socket URL → TCP via Auth Proxy.
# postgres://user:pass@/db?host=/cloudsql/INSTANCE  →  postgres://user:pass@127.0.0.1:PORT/db?sslmode=disable
DATABASE_URL="$(python3 - <<'PY' "$RAW_URL" "$PROXY_PORT"
import sys, urllib.parse
raw, port = sys.argv[1], sys.argv[2]
u = urllib.parse.urlparse(raw)
qs = urllib.parse.parse_qs(u.query)
qs.pop("host", None)
qs["sslmode"] = ["disable"]
userinfo = ""
if u.username:
    userinfo = urllib.parse.quote(u.username, safe="")
    if u.password is not None:
        userinfo += ":" + urllib.parse.quote(u.password, safe="")
    userinfo += "@"
path = u.path if u.path else "/petsfollow"
query = urllib.parse.urlencode({k: v[0] for k, v in qs.items()})
print(f"postgres://{userinfo}127.0.0.1:{port}{path}?{query}")
PY
)"

echo "→ Start cloud-sql-proxy on :${PROXY_PORT} (${CLOUDSQL_INSTANCE})"
"$PROXY_BIN" --address 127.0.0.1 --port "$PROXY_PORT" "$CLOUDSQL_INSTANCE" &
PROXY_PID=$!
cleanup_proxy() {
  kill "$PROXY_PID" 2>/dev/null || true
  wait "$PROXY_PID" 2>/dev/null || true
}
trap cleanup_proxy EXIT

PROXY_READY=false
for _ in $(seq 1 40); do
  if (echo >/dev/tcp/127.0.0.1/"$PROXY_PORT") >/dev/null 2>&1; then
    PROXY_READY=true
    break
  fi
  sleep 0.5
done
if [[ "$PROXY_READY" != true ]]; then
  echo "cloud-sql-proxy not ready on 127.0.0.1:${PROXY_PORT}" >&2
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "psql required (postgresql-client)" >&2
  exit 1
fi

# Marqueur obligatoire pour l'allowlist localhost après rewrite TCP.
export PF_CLEANUP_TARGET=staging
export DATABASE_URL
bash "$ROOT/scripts/cleanup-staging-quality.sh" "${CLEANUP_ARGS[@]+"${CLEANUP_ARGS[@]}"}"
