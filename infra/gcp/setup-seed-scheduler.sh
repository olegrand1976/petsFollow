#!/usr/bin/env bash
# OBSOLÈTE — le seed staging n’est plus planifié automatiquement.
# Voir delete-seed-scheduler.sh / make gcp-delete-seed-scheduler.
#
# Reset manuel :
#   - Admin Pro → zone danger (confirm « RESET STAGING »)
#   - bash infra/gcp/postdeploy.sh --seed
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "⚠ setup-seed-scheduler.sh est obsolète (plus de reset hebdo auto)." >&2
echo "→ Suppression du job Scheduler s’il existe…" >&2
exec bash "${SCRIPT_DIR}/delete-seed-scheduler.sh"
