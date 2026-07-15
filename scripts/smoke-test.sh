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
PET_ID=$(echo "$PETS" | python3 -c "import sys,json; d=json.load(sys.stdin)['data']; p=d[0] if d else {}; print(p.get('id') or p.get('ID',''))")

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
  -H "Authorization: Bearer $CLIENT_TOKEN" >/dev/null

curl -sf "$API/api/v1/pets/$PET_ID/timeline" -H "Authorization: Bearer $CLIENT_TOKEN" >/dev/null

# Auth register → confirm → forgot → reset (emails uniques)
AUTH_EMAIL="smoke+$(date +%s)@petsfollow.test"
REG=$(curl -sf -X POST "$API/api/v1/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$AUTH_EMAIL\",\"password\":\"SmokePass123!\",\"fullName\":\"Smoke Vet\",\"practiceName\":\"Smoke Practice\"}")
CONFIRM_PATH=$(echo "$REG" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['confirmPath'])")
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

echo "OK — smoke MVP + billing + auth reset passed"
