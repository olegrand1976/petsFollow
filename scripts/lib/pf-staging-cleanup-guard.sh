#!/usr/bin/env bash
# Garde-fou commun : cleanup staging uniquement.
# Source : source "$(dirname "$0")/lib/pf-staging-cleanup-guard.sh"
# Puis : pf_assert_staging_cleanup_url "$DATABASE_URL" "PF_ALLOW_PROD_QUALITY_CLEANUP"
#
# Autorisé :
#   - URL contenant premedica-db-staging
#   - localhost/127.0.0.1 SI PF_CLEANUP_TARGET=staging (posé par le wrapper GCP
#     après assert sur le secret / instance — évite le footgun proxy→prod)
# Override ops : $2=1 (ex. PF_ALLOW_PROD_QUALITY_CLEANUP=1)
set -euo pipefail

pf_assert_staging_cleanup_url() {
  local u="${1:-}"
  local override_var="${2:-PF_ALLOW_PROD_CLEANUP}"
  local override_val="${!override_var:-}"

  if [[ -z "$u" ]]; then
    echo "DATABASE_URL required" >&2
    return 1
  fi

  if [[ "$override_val" == "1" ]]; then
    echo "WARNING: ${override_var}=1 — allowlist bypassed" >&2
    return 0
  fi

  if [[ "$u" == *"petsfollow-db-prod"* ]]; then
    echo "Refusing DATABASE_URL targeting petsfollow-db-prod." >&2
    echo "Override only if intentional: ${override_var}=1" >&2
    return 1
  fi

  if [[ "$u" == *"premedica-db-staging"* ]]; then
    return 0
  fi

  # Proxy TCP rewrite : instance absente de l'URL — exiger le marqueur explicite.
  if [[ "$u" == *"@127.0.0.1:"* || "$u" == *"@localhost:"* ]]; then
    if [[ "${PF_CLEANUP_TARGET:-}" == "staging" ]]; then
      return 0
    fi
    echo "Refusing localhost DATABASE_URL without PF_CLEANUP_TARGET=staging." >&2
    echo "Le wrapper GCP pose ce flag après assert instance staging." >&2
    echo "Local Docker : PF_CLEANUP_TARGET=staging DATABASE_URL=… ./scripts/…" >&2
    echo "Override prod (ops) : ${override_var}=1" >&2
    return 1
  fi

  echo "Refusing DATABASE_URL outside staging allowlist (premedica-db-staging | localhost+PF_CLEANUP_TARGET=staging)." >&2
  echo "Override only if intentional: ${override_var}=1" >&2
  return 1
}
