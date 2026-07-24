#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
if [ -f "$ROOT/.env" ]; then set -a && source "$ROOT/.env" && set +a; fi
API="${PETSFOLLOW_API_URL:-http://localhost:${PETSFOLLOW_API_PORT:-8291}}"

echo "== petsFollow smoke ($API) =="

curl -sf "$API/health" | grep -q ok
curl -sf "$API/ready" | grep -q ready

VET=$(curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"vet.demo@petsfollow.test","password":"VetDemo123!"}')
VET_TOKEN=$(echo "$VET" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")
VET_REFRESH=$(echo "$VET" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['refreshToken'])")
REFRESHED=$(curl -sf -X POST "$API/api/v1/auth/refresh" -H 'Content-Type: application/json' \
  -d "{\"refreshToken\":\"$VET_REFRESH\"}")
echo "$REFRESHED" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; assert d.get('accessToken'), d"

CLIENT=$(curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"client.demo@petsfollow.test","password":"ClientDemo123!"}')
CLIENT_TOKEN=$(echo "$CLIENT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")

ADMIN=$(curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"admin.demo@petsfollow.test","password":"AdminDemo123!"}')
ADMIN_TOKEN=$(echo "$ADMIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")

curl -sf "$API/api/v1/clients" -H "Authorization: Bearer $VET_TOKEN" >/dev/null
curl -sf "$API/api/v1/billing/plans" >/dev/null
curl -sf "$API/api/v1/admin/metrics/overview" -H "Authorization: Bearer $ADMIN_TOKEN" >/dev/null

PETS=$(curl -sf "$API/api/v1/pets" -H "Authorization: Bearer $CLIENT_TOKEN")
PET_ID=$(echo "$PETS" | python3 -c "
import sys, json
from datetime import datetime, timezone
pets = json.load(sys.stdin).get('data') or []
now = datetime.now(timezone.utc)
for p in pets:
    if p.get('paymentStatus') != 'active':
        continue
    ent = p.get('entitlement') or {}
    if ent.get('status') not in ('active', 'past_due', 'cancelled'):
        continue
    vu = ent.get('validUntil')
    if vu:
        try:
            t = datetime.fromisoformat(vu.replace('Z', '+00:00'))
            if t.tzinfo is None:
                t = t.replace(tzinfo=timezone.utc)
            if t <= now:
                continue
        except Exception:
            pass
    print(p.get('id') or p.get('ID') or '')
    break
")

if [ -z "$PET_ID" ]; then
  CREATE=$(curl -sf -X POST "$API/api/v1/pets" -H "Authorization: Bearer $CLIENT_TOKEN" -H 'Content-Type: application/json' \
    -d '{"name":"SmokePet","species":"dog","breed":"test","plan":"triennial","billingMode":"subscription"}')
  PET_ID=$(echo "$CREATE" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; p=d.get('pet') or d; print(p.get('id') or p.get('ID',''))")
  OWNER_ID=$(echo "$CREATE" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; p=d.get('pet') or d; print(p.get('ownerUserId') or p.get('OwnerUserID',''))")
  curl -sf "$API/api/v1/billing/dev/mock-complete?pet_id=$PET_ID&owner_user_id=$OWNER_ID&plan_code=triennial&billing_mode=subscription" >/dev/null
fi

THREADS=$(curl -sf "$API/api/v1/messaging/threads" -H "Authorization: Bearer $CLIENT_TOKEN")
THREAD_ID=$(echo "$THREADS" | python3 -c "import sys,json; t=json.load(sys.stdin)['data'][0]; print(t.get('id') or t.get('ID',''))")
curl -sf -X POST "$API/api/v1/messaging/threads/$THREAD_ID/messages" \
  -H "Authorization: Bearer $CLIENT_TOKEN" -H 'Content-Type: application/json' \
  -d '{"body":"smoke test message"}' >/dev/null

SESS=$(curl -sf -X POST "$API/api/v1/pets/$PET_ID/heartrate/sessions" -H "Authorization: Bearer $CLIENT_TOKEN")
SESS_ID=$(echo "$SESS" | python3 -c "import sys,json; s=json.load(sys.stdin)['data']; print(s.get('id') or s.get('ID',''))")
curl -sf -X PATCH "$API/api/v1/heartrate/sessions/$SESS_ID" \
  -H "Authorization: Bearer $CLIENT_TOKEN" -H 'Content-Type: application/json' \
  -d '{"tapCount":60}' >/dev/null
curl -sf -X POST "$API/api/v1/heartrate/sessions/$SESS_ID/validate" \
  -H "Authorization: Bearer $CLIENT_TOKEN" -H 'Content-Type: application/json' \
  -d '{"comment":"smoke hr comment"}' >/dev/null

SESSIONS=$(curl -sf "$API/api/v1/pets/$PET_ID/heartrate/sessions" -H "Authorization: Bearer $CLIENT_TOKEN")
SESS_ID="$SESS_ID" python3 -c '
import json, os, sys
sess_id = os.environ["SESS_ID"]
rows = json.load(sys.stdin).get("data") or []
found = any((s.get("id") or s.get("ID")) == sess_id and (s.get("comment") or "") == "smoke hr comment" for s in rows)
if not found:
    raise SystemExit("heartrate comment missing after validate")
' <<<"$SESSIONS"

curl -sf "$API/api/v1/pets/$PET_ID/timeline" -H "Authorization: Bearer $CLIENT_TOKEN" >/dev/null

# Auth register → confirm → forgot → reset (emails uniques).
# confirmPath/resetPath ne sont exposés que si DEV_SEED_ENABLED=true (local/CI) ;
# sur staging/prod le register doit aboutir SANS fuiter les tokens → flux confirm/reset skippé.
AUTH_EMAIL="smoke+$(date +%s)@petsfollow.test"
REG=$(curl -sf -X POST "$API/api/v1/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$AUTH_EMAIL\",\"password\":\"SmokePass123!\",\"fullName\":\"Smoke Vet\",\"practiceName\":\"Smoke Practice\",\"consent\":true}")
CONFIRM_PATH=$(echo "$REG" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'].get('confirmPath') or '')")
if [ -n "$CONFIRM_PATH" ]; then
  CONFIRM_TOKEN=${CONFIRM_PATH#*token=}
  curl -sf -X POST "$API/api/v1/auth/confirm-email" -H 'Content-Type: application/json' \
    -d "{\"token\":\"$CONFIRM_TOKEN\"}" >/dev/null
  FORGOT=$(curl -sf -X POST "$API/api/v1/auth/forgot-password" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$AUTH_EMAIL\"}")
  RESET_PATH=$(echo "$FORGOT" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['resetPath'])")
  RESET_TOKEN=${RESET_PATH#*token=}
  curl -sf -X POST "$API/api/v1/auth/reset-password" -H 'Content-Type: application/json' \
    -d "{\"token\":\"$RESET_TOKEN\",\"password\":\"SmokePass456!\"}" >/dev/null
  curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$AUTH_EMAIL\",\"password\":\"SmokePass456!\"}" >/dev/null
else
  echo "  (confirmPath masqué — flux confirm/forgot/reset skippé, env non-dev)"
fi

# Register client + confirm
CLIENT_REG_EMAIL="smoke-client+$(date +%s)@petsfollow.test"
CLIENT_REG=$(curl -sf -X POST "$API/api/v1/auth/register-client" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$CLIENT_REG_EMAIL\",\"password\":\"ClientPass123!\",\"fullName\":\"Smoke Client\",\"consent\":true}")
CLIENT_CONFIRM_PATH=$(echo "$CLIENT_REG" | python3 -c "import sys,json; print(json.load(sys.stdin)['data'].get('confirmPath') or '')")
if [ -n "$CLIENT_CONFIRM_PATH" ]; then
  CLIENT_CONFIRM_TOKEN=${CLIENT_CONFIRM_PATH#*token=}
  curl -sf -X POST "$API/api/v1/auth/confirm-email" -H 'Content-Type: application/json' \
    -d "{\"token\":\"$CLIENT_CONFIRM_TOKEN\"}" >/dev/null
  curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$CLIENT_REG_EMAIL\",\"password\":\"ClientPass123!\"}" >/dev/null
fi

# Pet shares list (owner) + public media visit-reports blocked
curl -sf "$API/api/v1/pets/$PET_ID/shares" -H "Authorization: Bearer $CLIENT_TOKEN" >/dev/null
MEDIA_CODE=$(curl -s -o /dev/null -w '%{http_code}' "$API/media/visit-reports/smoke-forbidden.m4a")
# Public /media must not expose visit-reports (403 deny middleware or 404 if not on local FS/GCS).
test "$MEDIA_CODE" = "403" -o "$MEDIA_CODE" = "404"

COMM=$(curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"commercial.demo@petsfollow.test","password":"CommercialDemo123!"}')
COMM_TOKEN=$(echo "$COMM" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")
curl -sf "$API/api/v1/commercial/overview" -H "Authorization: Bearer $COMM_TOKEN" >/dev/null
curl -sf -X POST "$API/api/v1/admin/commercials" -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"email\":\"smoke-comm+$(date +%s)@petsfollow.test\",\"password\":\"CommercialDemo123!\",\"fullName\":\"Smoke Comm\"}" >/dev/null

MGR=$(curl -sf -X POST "$API/api/v1/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"commercial.manager@petsfollow.test","password":"CommercialDemo123!"}')
MGR_TOKEN=$(echo "$MGR" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['accessToken'])")
MGR_ME=$(curl -sf "$API/api/v1/me" -H "Authorization: Bearer $MGR_TOKEN")
echo "$MGR_ME" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; assert d.get('role')=='commercial_manager', d"
curl -sf "$API/api/v1/commercial-manager/overview" -H "Authorization: Bearer $MGR_TOKEN" >/dev/null

# Mauvais MDP → 401 (régression login)
BAD=$(curl -s -o /tmp/pf-smoke-bad-login.json -w '%{http_code}' -X POST "$API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"vet.demo@petsfollow.test","password":"WrongPass999!"}')
test "$BAD" = "401"

echo "OK — smoke MVP + billing + auth reset + register-client + shares/media + commercial + manager passed"
