#!/usr/bin/env bash
# Resend confirmation emails for unverified staging client profiles.
# Usage:
#   ./scripts/resend-staging-confirmations.sh email1@x.com email2@y.com
#   DATABASE_URL=... ./scripts/resend-staging-confirmations.sh --from-db
set -euo pipefail

API_BASE="${PETSFOLLOW_API_PUBLIC_URL:-https://api.petsfollow.ll-it-sc.be}"
API_BASE="${API_BASE%/}"

resend_one() {
  local email="$1"
  local code
  code="$(curl -sS -o /tmp/pf-resend.json -w '%{http_code}' \
    -X POST "${API_BASE}/api/v1/auth/resend-confirmation" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"${email}\"}")"
  if [[ "$code" != "200" ]]; then
    echo "FAIL ${email} HTTP ${code}: $(cat /tmp/pf-resend.json)" >&2
    return 1
  fi
  echo "OK   ${email}"
}

emails=()
if [[ "${1:-}" == "--from-db" ]]; then
  if [[ -z "${DATABASE_URL:-}" ]]; then
    echo "DATABASE_URL required with --from-db" >&2
    exit 1
  fi
  mapfile -t emails < <(psql "$DATABASE_URL" -Atc "
    SELECT email FROM identity.users
    WHERE role = 'client'
      AND email_verified_at IS NULL
      AND COALESCE(password_hash,'') <> ''
      AND email NOT LIKE '%@petsfollow.test'
    ORDER BY created_at ASC;")
else
  emails=("$@")
fi

if [[ ${#emails[@]} -eq 0 ]]; then
  echo "No emails to resend." >&2
  exit 0
fi

echo "→ Resending ${#emails[@]} confirmation(s) via ${API_BASE}"
fail=0
for email in "${emails[@]}"; do
  [[ -z "$email" ]] && continue
  resend_one "$email" || fail=$((fail + 1))
  sleep 0.4
done
exit "$fail"
