#!/usr/bin/env bash
# Seed démo PACS : upload testdata/pacs/demo-rx.dcm sur un pet actif de client.demo.
# Prérequis : make up-pacs · make seed · make api-dev (PACS_ORTHANC_URL câblé).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API="${PETSFOLLOW_API_URL:-${NUXT_PUBLIC_API_BASE:-http://localhost:8291}}"
API="${API%/}"
DCM="${PACS_DEMO_DCM:-$ROOT/testdata/pacs/demo-rx.dcm}"
SITE="${PETSFOLLOW_PUBLIC_SITE_URL:-http://localhost:3002}"

# curl JSON API — affiche body + status en cas d'échec (évite -sf opaque).
api_json() {
  local method="$1" path="$2"
  shift 2
  local tmp code
  tmp="$(mktemp)"
  code="$(curl -sS -o "$tmp" -w '%{http_code}' -X "$method" "${API}${path}" "$@" || true)"
  if [[ "$code" -lt 200 || "$code" -ge 300 ]]; then
    echo "ERREUR HTTP ${code} ${method} ${path}" >&2
    cat "$tmp" >&2 || true
    echo >&2
    rm -f "$tmp"
    return 1
  fi
  cat "$tmp"
  rm -f "$tmp"
}

if [[ ! -f "$DCM" ]]; then
  echo "→ Génération fixture DICOM"
  python3 "$ROOT/scripts/gen-minimal-dicom.py" -o "$DCM"
fi

echo "→ Login vet.demo @ ${API}"
LOGIN="$(api_json POST /api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"vet.demo@petsfollow.test","password":"VetDemo123!"}')"
TOK="$(python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["accessToken"])' <<<"$LOGIN")"

echo "→ Resolve pet client.demo"
CLIENTS="$(api_json GET /api/v1/clients -H "Authorization: Bearer $TOK")"
CLIENT_ID="$(python3 -c '
import json,sys
rows=json.load(sys.stdin)["data"]
c=next((x for x in rows if x.get("email")=="client.demo@petsfollow.test"), None)
assert c, "client.demo not found — make seed"
print(c["userId"])
' <<<"$CLIENTS")"

PETS="$(api_json GET "/api/v1/clients/$CLIENT_ID/pets" -H "Authorization: Bearer $TOK")"
read -r PET_ID PET_NAME <<<"$(python3 -c '
import json,sys
rows=json.load(sys.stdin)["data"]
p=next((x for x in rows if x.get("paymentStatus")=="active"), rows[0] if rows else None)
assert p, "no pet"
print(p["id"], p.get("name") or "pet")
' <<<"$PETS")"

echo "→ Wake PACS"
api_json POST /api/v1/pacs/wake \
  -H "Authorization: Bearer $TOK" \
  -H 'Content-Type: application/json' \
  -d '{}' >/dev/null || true

echo "→ Wait ready (max 60s)"
STATE=""
for _ in $(seq 1 30); do
  if ST="$(api_json GET /api/v1/pacs/status -H "Authorization: Bearer $TOK" 2>/dev/null)"; then
    STATE="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("data",{}).get("state",""))' <<<"$ST" 2>/dev/null || true)"
    if [[ "$STATE" == "ready" ]]; then
      break
    fi
  fi
  sleep 2
done
if [[ "${STATE:-}" != "ready" ]]; then
  echo "ERREUR: PACS non ready (state=${STATE:-unknown}). make up-pacs + PACS_ORTHANC_URL ?" >&2
  exit 1
fi

echo "→ Upload $DCM → pet ${PET_NAME} ($PET_ID)"
UP="$(api_json POST "/api/v1/pets/$PET_ID/pacs/studies" \
  -H "Authorization: Bearer $TOK" \
  -F "file=@${DCM};type=application/dicom" \
  -F "description=RX thorax demo petsFollow")"
STUDY="$(python3 -c 'import json,sys; d=json.load(sys.stdin)["data"]; print(d.get("orthancStudyId",""), d.get("id",""))' <<<"$UP")"

echo ""
echo "OK — étude liée : $STUDY"
echo "Ouvrir : ${SITE}/clients/${CLIENT_ID}/pets/${PET_ID}?tab=imaging"
