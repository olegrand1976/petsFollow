#!/usr/bin/env bash
# Smoke eID BE — Viewer import + Web eID challenge/origin + verify bad token.
# Local :  make smoke-eid
# Staging : make smoke-eid-staging
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
if [ -f "$ROOT/.env" ]; then set -a && source "$ROOT/.env" && set +a; fi

API_RAW="${PETSFOLLOW_API_URL:-http://localhost:${PETSFOLLOW_API_PORT:-8291}}"
API="${API_RAW%/}/api/v1"
API="${API%/api/v1}/api/v1"

# Staging default ; override locally via EID_SITE_ORIGIN / EID_SITE_ORIGIN_EXPECT.
if [[ "$API" == *"petsfollow.ll-it-sc.be"* ]]; then
  SITE_EXPECT="${EID_SITE_ORIGIN_EXPECT:-https://petsfollow.ll-it-sc.be}"
else
  SITE_EXPECT="${EID_SITE_ORIGIN_EXPECT:-${EID_SITE_ORIGIN:-http://localhost:3002}}"
fi

EMAIL="${EID_SMOKE_EMAIL:-vet.demo@petsfollow.test}"
PASS="${EID_SMOKE_PASS:-VetDemo123!}"
FIXTURE="$ROOT/go/internal/eid/testdata/sample_valid.eid"
SECRET="${BFF_PROXY_SECRET:-}"

echo "== petsFollow eID smoke ($API) =="
echo "→ Login $EMAIL"
LOGIN=$(curl -sS -X POST "$API/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASS\"}")
TOKEN=$(printf '%s' "$LOGIN" | python3 -c 'import json,sys; d=json.load(sys.stdin); print((d.get("data") or d).get("accessToken") or (d.get("data") or d).get("token") or "")')
if [[ -z "$TOKEN" ]]; then
  echo "login failed: $LOGIN" >&2
  exit 1
fi
echo "  login ok"

echo "→ POST /vet/eid/import (Viewer)"
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

echo "→ POST /vet/eid/web-eid/verify (bad token → 422, burns nonce)"
VER=$(curl -sS -w '\n%{http_code}' -X POST "$API/vet/eid/web-eid/verify" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"token":{"unverifiedCertificate":"smoke-bad"}}')
VER_BODY=$(printf '%s' "$VER" | sed '$d')
VER_CODE=$(printf '%s' "$VER" | tail -n1)
echo "  HTTP $VER_CODE"
[[ "$VER_CODE" == "422" ]] || { echo "want 422 got $VER_CODE $VER_BODY" >&2; exit 1; }
MSG=$(printf '%s' "$VER_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin); err=d.get("error") or {}; print(err.get("msgKey") or err.get("messageKey") or err.get("message") or "")')
[[ "$MSG" != "eid_nonce_expired" ]] || { echo "nonce burned before validate: $VER_BODY" >&2; exit 1; }

echo "→ POST /vet/eid/web-eid/verify again (nonce consumed → eid_nonce_expired)"
VER2=$(curl -sS -w '\n%{http_code}' -X POST "$API/vet/eid/web-eid/verify" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"token":{"unverifiedCertificate":"smoke-bad"}}')
VER2_BODY=$(printf '%s' "$VER2" | sed '$d')
VER2_CODE=$(printf '%s' "$VER2" | tail -n1)
echo "  HTTP $VER2_CODE"
[[ "$VER2_CODE" == "422" ]] || { echo "want 422 got $VER2_CODE $VER2_BODY" >&2; exit 1; }
printf '%s\n' "$VER2_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin); err=d.get("error") or {}; msg=err.get("msgKey") or err.get("messageKey") or ""; assert msg=="eid_nonce_expired", d; print("  nonce expired OK")'

if [[ -n "$SECRET" ]]; then
  # Loopback alias only when configured origin is localhost-family.
  case "$SITE_EXPECT" in
    http://localhost:*|http://127.0.0.1:*)
      ALIAS="http://127.0.0.1:${SITE_EXPECT##*:}"
      if [[ "$SITE_EXPECT" == http://127.0.0.1:* ]]; then
        ALIAS="http://localhost:${SITE_EXPECT##*:}"
      fi
      echo "→ GET challenge with X-PF-Web-Eid-Origin=$ALIAS (BFF secret)"
      CH2=$(curl -sS -w '\n%{http_code}' -X GET "$API/vet/eid/web-eid/challenge" \
        -H "Authorization: Bearer $TOKEN" \
        -H "X-PF-Proxy-Secret: $SECRET" \
        -H "X-PF-Web-Eid-Origin: $ALIAS")
      CH2_BODY=$(printf '%s' "$CH2" | sed '$d')
      CH2_CODE=$(printf '%s' "$CH2" | tail -n1)
      ORIGIN2=$(printf '%s' "$CH2_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin); data=d.get("data") or d; print(data.get("origin") or "")')
      echo "  HTTP $CH2_CODE origin=$ORIGIN2"
      [[ "$CH2_CODE" == "200" ]]
      [[ "$ORIGIN2" == "$ALIAS" ]] || { echo "want alias $ALIAS got $ORIGIN2" >&2; exit 1; }

      echo "→ GET challenge forged origin without secret (ignored)"
      CH3=$(curl -sS -w '\n%{http_code}' -X GET "$API/vet/eid/web-eid/challenge" \
        -H "Authorization: Bearer $TOKEN" \
        -H "X-PF-Web-Eid-Origin: $ALIAS")
      CH3_BODY=$(printf '%s' "$CH3" | sed '$d')
      ORIGIN3=$(printf '%s' "$CH3_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin); data=d.get("data") or d; print(data.get("origin") or "")')
      echo "  origin=$ORIGIN3 (expect configured $SITE_EXPECT)"
      [[ "$ORIGIN3" == "$SITE_EXPECT" ]] || {
        echo "forged origin accepted without secret (got $ORIGIN3, want $SITE_EXPECT)" >&2
        echo "API outdated — restart make api-dev to pick up X-PF-Web-Eid-Origin proxy gate" >&2
        exit 1
      }
      ;;
  esac
fi

echo "OK — eID Viewer + Web eID challenge/verify contract"
