#!/usr/bin/env bash
# Smoke against shared crewai-orchestrator.
# Skip cleanly when CREWAI_BASE_URL is empty.
set -euo pipefail

BASE="${CREWAI_BASE_URL:-}"
SECRET="${CREWAI_SHARED_SECRET:-}"

if [[ -z "$BASE" ]]; then
  echo "crewai-smoke: SKIP (CREWAI_BASE_URL empty)"
  exit 0
fi

BASE="${BASE%/}"

mint_id_token() {
  # SA accounts support --audiences; user accounts do not.
  local tok=""
  tok="$(gcloud auth print-identity-token --audiences="$BASE" 2>/dev/null || true)"
  if [[ -n "$tok" ]]; then
    printf '%s' "$tok"
    return 0
  fi
  if [[ -n "${CREWAI_IMPERSONATE_SA:-}" ]]; then
    tok="$(gcloud auth print-identity-token \
      --impersonate-service-account="$CREWAI_IMPERSONATE_SA" \
      --audiences="$BASE" 2>/dev/null || true)"
    if [[ -n "$tok" ]]; then
      printf '%s' "$tok"
      return 0
    fi
  fi
  # User ADC token — works for Cloud Run invoker when the user has roles/run.invoker.
  tok="$(gcloud auth print-identity-token 2>/dev/null || true)"
  printf '%s' "$tok"
}

hdr=(-H "Content-Type: application/json" -H "X-Crew-API-Version: 1")
if [[ -n "$SECRET" ]]; then
  hdr+=(-H "X-Crew-Secret: $SECRET")
fi
if [[ "${CREWAI_USE_ID_TOKEN:-}" == "true" || "${CREWAI_USE_ID_TOKEN:-}" == "1" ]]; then
  tok="$(mint_id_token)"
  if [[ -z "$tok" ]]; then
    echo "crewai-smoke: CREWAI_USE_ID_TOKEN set but cannot mint identity token" >&2
    exit 1
  fi
  hdr+=(-H "Authorization: Bearer $tok")
fi

echo "→ health $BASE/health"
code="$(curl -sS -o /tmp/crewai-health.json -w '%{http_code}' "${hdr[@]}" "$BASE/health")"
if [[ "$code" != "200" ]]; then
  echo "health failed: HTTP $code" >&2
  cat /tmp/crewai-health.json >&2 || true
  exit 1
fi

echo "→ submit staging_smoke_test"
code="$(curl -sS -o /tmp/crewai-smoke.json -w '%{http_code}' "${hdr[@]}" \
  -d '{"workflow":"staging_smoke_test","tenant":"platform","correlationId":"smoke-local","payload":{"ping":true}}' \
  "$BASE/v1/tasks/submit")"
if [[ "$code" != "200" ]]; then
  echo "smoke submit failed: HTTP $code" >&2
  cat /tmp/crewai-smoke.json >&2 || true
  exit 1
fi
python3 - <<'PY'
import json
d=json.load(open("/tmp/crewai-smoke.json"))
assert d.get("status")=="completed", d
print("smoke ok", d.get("taskId"))
PY

echo "→ submit petsfollow_cr_improve (fixture)"
code="$(curl -sS -o /tmp/crewai-cr.json -w '%{http_code}' "${hdr[@]}" \
  -d '{"workflow":"petsfollow_cr_improve","tenant":"petsfollow","correlationId":"smoke-cr","payload":{"sourceText":"Chien toux 3j, T38.5","targetLocale":"fr","countryCode":"BE"}}' \
  "$BASE/v1/tasks/submit")"
if [[ "$code" != "200" ]]; then
  echo "cr improve submit failed: HTTP $code" >&2
  cat /tmp/crewai-cr.json >&2 || true
  exit 1
fi
python3 - <<'PY'
import json
d=json.load(open("/tmp/crewai-cr.json"))
assert d.get("status")=="completed", d
md=(d.get("result") or {}).get("markdown") or ""
assert "Anamnèse" in md or "Anamnese" in md or "History" in md, md[:200]
assert len(d.get("steps") or []) >= 3, d.get("steps")
print("cr improve ok", d.get("taskId"))
PY

echo "crewai-smoke: OK"
