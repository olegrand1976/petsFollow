#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/gcp-env.sh"
echo "→ postdeploy: migrate + seed-reset + smoke on ${API_CUSTOM_DOMAIN}"
