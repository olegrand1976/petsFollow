#!/usr/bin/env bash
# Smoke Flux A — 1 cabinet via Billit master (live sandbox).
# ⚠ Flux A est EN SOMMEIL : l'API répond 404 invoicing_saas_disabled tant que
# INVOICING_SAAS_ENABLED=true n'est pas posé sur l'API. Script conservé pour le
# jour où le flux est réveillé.
# Prérequis : API live (mock off) + INVOICING_SAAS_ENABLED=true + BILLIT_MASTER_*
# + seed + saas_billing_enabled.
# Usage :
#   BILLIT_MASTER_PARTY_ID=… BILLIT_MASTER_API_KEY=… make billit-saas-master-smoke
# Optionnel :
#   BILLIT_SMOKE_SAAS_PRACTICE_ID=<uuid>   # sinon 1er target opt-in
#   BILLIT_SMOKE_SAAS_SEND=1              # envoi Peppol après draft
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
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
for _k in BILLIT_ENABLED BILLIT_MOCK_ENABLED BILLIT_MASTER_PARTY_ID BILLIT_MASTER_API_KEY \
  BILLIT_SMOKE_SAAS_PRACTICE_ID BILLIT_SMOKE_SAAS_SEND PETSFOLLOW_API_URL PETSFOLLOW_API_PORT; do
  _preserve "$_k"
done
if [ -f "$ROOT/.env" ]; then set -a && source "$ROOT/.env" && set +a; fi
for _k in BILLIT_ENABLED BILLIT_MOCK_ENABLED BILLIT_MASTER_PARTY_ID BILLIT_MASTER_API_KEY \
  BILLIT_SMOKE_SAAS_PRACTICE_ID BILLIT_SMOKE_SAAS_SEND PETSFOLLOW_API_URL PETSFOLLOW_API_PORT; do
  _restore "$_k"
done

API="${PETSFOLLOW_API_URL:-http://localhost:${PETSFOLLOW_API_PORT:-8291}}"
API="${API%/}"

red() { printf '\033[31m%s\033[0m\n' "$*" >&2; }
ok() { printf '  OK  %s\n' "$*"; }
fail() { red "FAIL  $*"; exit 1; }

echo "== Billit SaaS master smoke ($API) =="

if [ "${BILLIT_ENABLED:-}" != "true" ] && [ "${BILLIT_ENABLED:-}" != "1" ]; then
  fail "BILLIT_ENABLED must be true"
fi
if [ "${BILLIT_MOCK_ENABLED:-}" = "true" ] || [ "${BILLIT_MOCK_ENABLED:-}" = "1" ]; then
  fail "BILLIT_MOCK_ENABLED is true — refuse live master smoke"
fi
MASTER_PARTY="${BILLIT_MASTER_PARTY_ID:-}"
MASTER_KEY="${BILLIT_MASTER_API_KEY:-}"
if [ -z "$MASTER_PARTY" ] || [ -z "$MASTER_KEY" ]; then
  fail "BILLIT_MASTER_PARTY_ID + BILLIT_MASTER_API_KEY required"
fi
ok "env live master credentials"

curl -sf "$API/health" | grep -q ok || fail "API /health"
ok "API up"

ADMIN=$(curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"admin.demo@petsfollow.test","password":"AdminDemo123!"}') \
  || fail "login admin.demo (seed required)"
ADMIN_TOKEN=$(echo "$ADMIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")

TARGETS=$(curl -sf "$API/api/v1/admin/invoicing/saas-targets" \
  -H "Authorization: Bearer $ADMIN_TOKEN") || fail "GET saas-targets"
PRACTICE="${BILLIT_SMOKE_SAAS_PRACTICE_ID:-}"
if [ -z "$PRACTICE" ]; then
  PRACTICE=$(echo "$TARGETS" | python3 -c "
import sys, json
rows = json.load(sys.stdin).get('data') or []
for r in rows:
    if r.get('saasBillingEnabled') and r.get('saasDraftEnabled'):
        print(r['practiceId']); break
")
fi
[ -n "$PRACTICE" ] || fail "no opted-in SaaS target (enable Flux A on /admin/invoicing or set BILLIT_SMOKE_SAAS_PRACTICE_ID)"
ok "practice $PRACTICE"

# Ensure opt-in (idempotent)
curl -sf -X POST "$API/api/v1/admin/invoicing/practices/$PRACTICE/saas-billing" \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"enabled":true}' >/tmp/pf-saas-enable.json \
  || fail "saas-billing enable (practice must be active BE fiscal)"
ok "saas_billing_enabled"

DRAFT=$(curl -sf -X POST "$API/api/v1/admin/invoicing/connections/$PRACTICE/saas-draft" \
  -H "Authorization: Bearer $ADMIN_TOKEN") || fail "saas-draft (Billit CreateDocument master)"
DOC_ID=$(echo "$DRAFT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
ORDER=$(echo "$DRAFT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'].get('billitOrderId') or '')")
SRC=$(echo "$DRAFT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'].get('source'))")
[ "$SRC" = "saas_master" ] || fail "want source=saas_master got $SRC"
[ -n "$ORDER" ] || fail "empty billitOrderId after draft"
ok "draft $DOC_ID order=$ORDER"

if [ "${BILLIT_SMOKE_SAAS_SEND:-}" = "1" ] || [ "${BILLIT_SMOKE_SAAS_SEND:-}" = "true" ]; then
  SEND=$(curl -sf -X POST "$API/api/v1/admin/invoicing/connections/$PRACTICE/saas-documents/$DOC_ID/send" \
    -H "Authorization: Bearer $ADMIN_TOKEN") || fail "saas send Peppol master"
  STATUS=$(echo "$SEND" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'].get('status'))")
  case "$STATUS" in
    sending|delivered) ok "send → status=$STATUS" ;;
    *) fail "send unexpected status=$STATUS" ;;
  esac
  echo "== Billit SaaS master smoke: PASS (draft+send) =="
  echo "Attendre webhook delivered pour order $ORDER."
else
  echo "== Billit SaaS master smoke: PASS (draft only) =="
  echo "Pour send Peppol : BILLIT_SMOKE_SAAS_SEND=1 make billit-saas-master-smoke"
fi
