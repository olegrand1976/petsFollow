#!/usr/bin/env bash
# Smoke S6 pharmacie : settings → search → receipt → DAF finalize → movements (+ VAMReg dry-run si antibio).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
if [ -f "$ROOT/.env" ]; then set -a && source "$ROOT/.env" && set +a; fi
API="${PETSFOLLOW_API_URL:-http://localhost:${PETSFOLLOW_API_PORT:-8291}}"

echo "== petsFollow pharmacy S6 smoke ($API) =="

python3 - "$API" <<'PY'
import json, sys, urllib.request, urllib.error, datetime, uuid

API = sys.argv[1]
suffix = uuid.uuid4().hex[:8]

def req(method, path, token=None, body=None):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    r = urllib.request.Request(API + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(r, timeout=60) as resp:
            raw = resp.read().decode()
            return resp.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        raw = e.read().decode()
        try:
            env = json.loads(raw) if raw else {}
        except Exception:
            env = {"raw": raw}
        return e.code, env

code, login = req("POST", "/api/v1/auth/login", body={
    "email": "vet.demo@petsfollow.test", "password": "VetDemo123!",
})
if code != 200:
    raise SystemExit(f"login failed {code} {login}")
tok = login["data"]["accessToken"]
print("login ok")

code, env = req("GET", "/api/v1/vet/pharmacy/settings", tok)
if code != 200:
    raise SystemExit(f"settings {code} {env}")
print("settings ok")

code, env = req("GET", "/api/v1/vet/pharmacy/medications/search?q=2712345&limit=5", tok)
if code != 200:
    raise SystemExit(f"search {code} {env} (pg_trgm/search broken?)")
items = (env.get("data") or {}).get("items") or []
if not items:
    code, env = req("GET", "/api/v1/vet/pharmacy/medications/search?q=a&limit=10", tok)
    items = (env.get("data") or {}).get("items") or []
if not items:
    raise SystemExit("FAIL: catalogue médicaments vide — seed / import-cnk requis")
med = items[0]
med_id = med["id"]
print("med", med.get("cnk"), med.get("name"), "ab=", med.get("isAntibiotic"))

exp = (datetime.date.today() + datetime.timedelta(days=120)).isoformat()
code, env = req("POST", "/api/v1/vet/pharmacy/batches", tok, {
    "medicationId": med_id, "lotNumber": f"S6-{suffix}", "expiresOn": exp, "qty": 3,
})
if code not in (200, 201):
    raise SystemExit(f"receipt {code} {env}")
print("receipt ok")

item = {"medicationId": med_id, "qty": 1, "ammNumber": f"BE-S6-{suffix}"}
if med.get("isAntibiotic"):
    item["vamregPayload"] = {"species": "dog", "indication": "smoke", "durationDays": 3, "posology": "1x/j"}

code, env = req("POST", "/api/v1/vet/pharmacy/daf", tok, {"items": [item]})
if code != 201:
    raise SystemExit(f"draft {code} {env}")
daf_id = env["data"]["id"]
print("draft", daf_id)

code, env = req("POST", f"/api/v1/vet/pharmacy/daf/{daf_id}/finalize", tok)
if code != 200:
    raise SystemExit(f"finalize {code} {env}")
doc = env.get("data") or {}
if isinstance(doc.get("daf"), dict):
    doc = doc["daf"]
print("finalize", doc.get("status"), "vamreg=", doc.get("vamregStatus"), "n=", doc.get("displayNumber"))

code, env = req("GET", f"/api/v1/vet/pharmacy/movements?dafId={daf_id}&limit=10", tok)
if code != 200:
    raise SystemExit(f"movements {code} {env}")
mov = (env.get("data") or {}).get("items") or []
if not any(m.get("dafId") == daf_id for m in mov):
    raise SystemExit(f"no daf-linked movements {env}")
print("movements ok", len(mov))
print("S6_PHARMACY_SMOKE_OK")
PY
