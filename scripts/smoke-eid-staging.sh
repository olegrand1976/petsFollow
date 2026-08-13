#!/usr/bin/env bash
# Compat : délègue au smoke eID générique contre staging.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
exec env \
  PETSFOLLOW_API_URL="${PETSFOLLOW_API_URL:-https://api.petsfollow.ll-it-sc.be}" \
  EID_SITE_ORIGIN_EXPECT="${EID_SITE_ORIGIN_EXPECT:-https://petsfollow.ll-it-sc.be}" \
  bash "$ROOT/scripts/smoke-eid.sh"
