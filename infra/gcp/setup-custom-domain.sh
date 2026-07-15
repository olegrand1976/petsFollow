#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/gcp-env.sh"
DOMAIN="${CUSTOM_DOMAIN:-petsfollow.ll-it-sc.be}"
echo "→ Configure custom domain ${DOMAIN} (pattern Kore setup-custom-domain.sh)"
echo "  1. gcloud compute network-endpoint-groups create ${FRONTEND_SERVICE}-neg ..."
echo "  2. DNS OVH: ${DOMAIN} → LB IP"
echo "  3. API: ${API_CUSTOM_DOMAIN} → same LB with host rule to ${API_SERVICE}"
