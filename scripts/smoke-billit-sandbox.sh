#!/usr/bin/env bash
# Smoke Billit sandbox (live) — checklist 34 A–F (partie automatisable).
# Prérequis : API joignable avec BILLIT_MOCK_ENABLED=false + secrets live.
# Usage :
#   BILLIT_WEBHOOK_SECRET=… make billit-sandbox-smoke
# Optionnel (parcours send) :
#   BILLIT_SMOKE_PARTY_ID=… BILLIT_SMOKE_API_KEY=… make billit-sandbox-smoke
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# Preserve caller overrides (make/env) so .env mock=true cannot hide a live smoke.
_preserve() {
  local n="$1"
  eval "_had_$n=0"
  if eval "[ \"\${$n+x}\" ]"; then
    eval "_had_$n=1"
    eval "_val_$n=\"\${$n}\""
  fi
}
_restore() {
  local n="$1"
  if eval "[ \"\$_had_$n\" = 1 ]"; then
    eval "export $n=\"\$_val_$n\""
  fi
}
for _k in BILLIT_ENABLED BILLIT_MOCK_ENABLED BILLIT_WEBHOOK_SECRET BILLIT_SECRETS_BACKEND BILLIT_SECRETS_KEY \
  BILLIT_SMOKE_PARTY_ID BILLIT_SMOKE_API_KEY PETSFOLLOW_API_URL PETSFOLLOW_API_PORT; do
  _preserve "$_k"
done
if [ -f "$ROOT/.env" ]; then set -a && source "$ROOT/.env" && set +a; fi
for _k in BILLIT_ENABLED BILLIT_MOCK_ENABLED BILLIT_WEBHOOK_SECRET BILLIT_SECRETS_BACKEND BILLIT_SECRETS_KEY \
  BILLIT_SMOKE_PARTY_ID BILLIT_SMOKE_API_KEY PETSFOLLOW_API_URL PETSFOLLOW_API_PORT; do
  _restore "$_k"
done

API="${PETSFOLLOW_API_URL:-http://localhost:${PETSFOLLOW_API_PORT:-8291}}"
API="${API%/}"

red() { printf '\033[31m%s\033[0m\n' "$*" >&2; }
ok() { printf '  OK  %s\n' "$*"; }
fail() { red "FAIL  $*"; exit 1; }
skip() { printf '  SKIP %s\n' "$*"; }

echo "== Billit sandbox smoke ($API) =="

# --- Préflight config locale / process ---
MOCK="${BILLIT_MOCK_ENABLED:-}"
ENABLED="${BILLIT_ENABLED:-}"
SECRET="${BILLIT_WEBHOOK_SECRET:-}"
BACKEND="${BILLIT_SECRETS_BACKEND:-}"

if [ "${ENABLED}" != "true" ] && [ "${ENABLED}" != "1" ]; then
  fail "BILLIT_ENABLED must be true (got '${ENABLED:-empty}')"
fi
if [ "${MOCK}" = "true" ] || [ "${MOCK}" = "1" ]; then
  fail "BILLIT_MOCK_ENABLED is true — refuse smoke live. Use: BILLIT_MOCK_ENABLED=false make api-billit-live"
fi
if [ -z "${SECRET}" ]; then
  fail "BILLIT_WEBHOOK_SECRET required for live webhook checks"
fi
if [ "${BACKEND}" = "plain_dev" ] || [ -z "${BACKEND}" ]; then
  fail "BILLIT_SECRETS_BACKEND=plain_dev refused for live smoke (use local_enc + BILLIT_SECRETS_KEY)"
fi
if [ -z "${BILLIT_SECRETS_KEY:-}" ]; then
  fail "BILLIT_SECRETS_KEY required when BILLIT_SECRETS_BACKEND=local_enc"
fi
ok "env live (mock off, webhook secret, local_enc)"

# --- Health ---
curl -sf "$API/health" | grep -q ok || fail "API /health"
ok "API up"

# --- C. Webhook signature ---
CODE=$(curl -sS -o /tmp/pf-billit-wh.json -w '%{http_code}' \
  -X POST "$API/api/v1/invoicing/webhooks/billit" \
  -H 'Content-Type: application/json' \
  -d '{"OrderID":1,"Status":"delivered"}')
[ "$CODE" = "401" ] || fail "unsigned webhook want 401 got $CODE ($(cat /tmp/pf-billit-wh.json))"
ok "webhook unsigned → 401"

BODY='{"OrderID":999002,"EventType":"OrderDelivered","Status":"delivered","EventID":"smoke-unknown-order"}'
SIG=$(SECRET="$SECRET" BODY="$BODY" python3 - <<'PY'
import hmac, hashlib, os
secret = os.environ["SECRET"].encode()
body = os.environ["BODY"].encode()
print(hmac.new(secret, body, hashlib.sha256).hexdigest())
PY
)
CODE=$(curl -sS -o /tmp/pf-billit-wh.json -w '%{http_code}' \
  -X POST "$API/api/v1/invoicing/webhooks/billit" \
  -H 'Content-Type: application/json' \
  -H "X-Billit-Signature: $SIG" \
  -d "$BODY")
if [ "$CODE" = "401" ]; then
  fail "signed webhook → 401 (BILLIT_WEBHOOK_SECRET mismatch vs running API — same secret as make api-billit-live)"
fi
[ "$CODE" = "503" ] || fail "unknown order want 503 got $CODE ($(cat /tmp/pf-billit-wh.json))"
ok "webhook signed unknown order → 503 (retryable)"

# --- Auth + routes ---
VET=$(curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"vet.demo@petsfollow.test","password":"VetDemo123!"}') \
  || fail "login vet.demo (seed required)"
VET_TOKEN=$(echo "$VET" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")

CODE=$(curl -sS -o /tmp/pf-billit-conn.json -w '%{http_code}' \
  "$API/api/v1/practices/me/invoicing/connection" \
  -H "Authorization: Bearer $VET_TOKEN")
[ "$CODE" = "200" ] || fail "GET connection want 200 got $CODE ($(cat /tmp/pf-billit-conn.json))"
ok "GET practice invoicing connection"

ADMIN=$(curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"admin.demo@petsfollow.test","password":"AdminDemo123!"}') \
  || fail "login admin.demo"
ADMIN_TOKEN=$(echo "$ADMIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")
CODE=$(curl -sS -o /tmp/pf-billit-admin.json -w '%{http_code}' \
  "$API/api/v1/admin/invoicing/connections" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
[ "$CODE" = "200" ] || fail "admin connections want 200 got $CODE"
ok "GET admin invoicing connections"

# --- D. Optionnel : connect + facture BE + send (Party/clé sandbox) ---
PARTY="${BILLIT_SMOKE_PARTY_ID:-}"
KEY="${BILLIT_SMOKE_API_KEY:-}"
if [ -z "$PARTY" ] || [ -z "$KEY" ]; then
  skip "BILLIT_SMOKE_PARTY_ID / BILLIT_SMOKE_API_KEY unset — skip connect/send live"
  echo "== Billit sandbox smoke: PASS (gates C + routes) =="
  echo "Suite manuelle : doc 34 checklist D–F (UI Pro + webhook delivered)."
  exit 0
fi

CONN_STATUS=$(python3 -c "import json; print(json.load(open('/tmp/pf-billit-conn.json')).get('data',{}).get('status',''))")
if [ "$CONN_STATUS" != "active" ] && [ "$CONN_STATUS" != "pending_kyc" ]; then
  START=$(curl -sf -X POST "$API/api/v1/practices/me/invoicing/connect/start" \
    -H "Authorization: Bearer $VET_TOKEN" -H 'Content-Type: application/json' -d '{}') \
    || fail "connect/start (profil practice incomplet ? legal/TVA/n°/email)"
  STATE=$(echo "$START" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['state'])")
  curl -sf -X POST "$API/api/v1/practices/me/invoicing/connect/complete" \
    -H "Authorization: Bearer $VET_TOKEN" -H 'Content-Type: application/json' \
    -d "{\"state\":\"$STATE\",\"partyId\":\"$PARTY\",\"apiKey\":\"$KEY\"}" >/tmp/pf-billit-complete.json \
    || fail "connect/complete (CheckParty Billit ?)"
  ok "connect complete → $(python3 -c "import json; print(json.load(open('/tmp/pf-billit-complete.json'))['data'].get('status'))")"
else
  ok "already connected ($CONN_STATUS)"
fi

CREATE=$(curl -sf -X POST "$API/api/v1/practices/me/invoicing/documents" \
  -H "Authorization: Bearer $VET_TOKEN" -H 'Content-Type: application/json' \
  -d "{
    \"type\": \"invoice\",
    \"counterparty\": {
      \"name\": \"Smoke BE $(date +%s)\",
      \"country\": \"BE\",
      \"vatNumber\": \"BE1000000021\",
      \"street\": \"Rue Smoke 1\",
      \"city\": \"Bruxelles\",
      \"postal\": \"1000\"
    },
    \"lines\": [{\"description\": \"Smoke consult\", \"quantity\": 1, \"unitPriceExclCents\": 4200, \"vatPercent\": 21}]
  }") || fail "create invoice BE"
DOC_ID=$(echo "$CREATE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
ok "created document $DOC_ID"

SEND=$(curl -sf -X POST "$API/api/v1/practices/me/invoicing/documents/$DOC_ID/send" \
  -H "Authorization: Bearer $VET_TOKEN") || fail "send invoice (Billit CreateDocument/SendPeppol)"
STATUS=$(echo "$SEND" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'].get('status'))")
ORDER=$(echo "$SEND" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'].get('billitOrderId') or '')")
[ -n "$ORDER" ] || fail "send returned empty billitOrderId"
case "$STATUS" in
  sending|delivered) ok "send → status=$STATUS order=$ORDER" ;;
  *) fail "send unexpected status=$STATUS (live expect sending until webhook)" ;;
esac

echo "== Billit sandbox smoke: PASS (gates + send) =="
echo "Attendre webhook delivered pour order $ORDER ; rejouer body → status=duplicate."
echo "Admin : /admin/invoicing — mark partner après bascule Billit."
