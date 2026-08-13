#!/usr/bin/env bash
# Smoke eID BE against staging (or PETSFOLLOW_API_URL). Viewer import + Web eID challenge origin.
set -euo pipefail
API="${PETSFOLLOW_API_URL:-https://api.petsfollow.ll-it-sc.be}/api/v1"
API="${API%/api/v1}/api/v1"
SITE_EXPECT="${EID_SITE_ORIGIN_EXPECT:-https://petsfollow.ll-it-sc.be}"
EMAIL="${EID_SMOKE_EMAIL:-vet.demo@petsfollow.test}"
PASS="${EID_SMOKE_PASS:-VetDemo123!}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FIXTURE="$ROOT/go/internal/eid/testdata/sample_valid.eid"

echo "→ Login $EMAIL @ $API"
LOGIN=$(curl -sS -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
TOKEN=$(printf '%s' "$LOGIN" | python3 -c 'import json,sys; d=json.load(sys.stdin); print((d.get("data") or d).get("accessToken") or (d.get("data") or d).get("token") or "")')
if [[ -z "$TOKEN" ]]; then
  echo "login failed: $LOGIN" >&2
  exit 1
fi

echo "→ POST /vet/eid/import"
IMP=$(curl -sS -w '\n%{http_code}' -X POST "$API/vet/eid/import" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@${FIXTURE};filename=sample_valid.eid;type=application/xml")
IMP_BODY=$(printf '%s' "$IMP" | sed '$d')
IMP_CODE=$(printf '%s' "$IMP" | tail -n1)
echo "  HTTP $IMP_CODE"
printf '%s\n' "$IMP_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin); data=d.get("data") or d; assert data.get("firstname")=="Camille" and data.get("lastname")=="Testeur", data; assert "photo_jpeg_base64" not in data, data; print("  identity OK", data.get("niss"), data.get("import_tool"))'
[[ "$IMP_CODE" == "200" ]]

echo "→ GET /vet/eid/web-eid/challenge"
CH=$(curl -sS -w '\n%{http_code}' -X GET "$API/vet/eid/web-eid/challenge" \
  -H "Authorization: Bearer $TOKEN")
CH_BODY=$(printf '%s' "$CH" | sed '$d')
CH_CODE=$(printf '%s' "$CH" | tail -n1)
echo "  HTTP $CH_CODE"
ORIGIN=$(printf '%s' "$CH_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin); data=d.get("data") or d; print(data.get("origin") or "")')
NONCE=$(printf '%s' "$CH_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin); data=d.get("data") or d; print(data.get("nonce") or "")')
echo "  origin=$ORIGIN"
echo "  nonce_len=${#NONCE}"
[[ "$CH_CODE" == "200" ]]
[[ -n "$NONCE" ]]
[[ "$ORIGIN" == "$SITE_EXPECT" ]] || { echo "origin want $SITE_EXPECT got $ORIGIN" >&2; exit 1; }

echo "OK — eID Viewer import + Web eID challenge (origin bound)"
