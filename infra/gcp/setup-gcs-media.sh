#!/usr/bin/env bash
# Crée le bucket GCS médias (avatars / photos animaux) + IAM.
# Usage: ./infra/gcp/setup-gcs-media.sh
#
# Note GCP : les conditions IAM sur allUsers sont refusées
# (PublicResourceAllowConditionCheck). La protection PHI visit-reports/
# repose donc sur l’app : Upload sans URL publique + stream auth + clés UUID.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

BUCKET="${GCS_MEDIA_BUCKET:-petsfollow-media}"
LOCATION="${GCS_MEDIA_LOCATION:-${GCP_RUN_REGION}}"
SA_EMAIL="${SERVICE_ACCOUNT}@${GCP_PROJECT_ID}.iam.gserviceaccount.com"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

echo "=== petsFollow GCS media — gs://${BUCKET} (${LOCATION}) ==="

if gcloud storage buckets describe "gs://${BUCKET}" --project="$GCP_PROJECT_ID" >/dev/null 2>&1; then
  echo "  Bucket gs://${BUCKET} existe déjà"
else
  gcloud storage buckets create "gs://${BUCKET}" \
    --project="$GCP_PROJECT_ID" \
    --location="$LOCATION" \
    --uniform-bucket-level-access \
    --no-public-access-prevention
  echo "  Bucket gs://${BUCKET} créé"
fi

echo "→ IAM objectAdmin pour ${SA_EMAIL}"
gcloud storage buckets add-iam-policy-binding "gs://${BUCKET}" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.objectAdmin" \
  --quiet >/dev/null

# Avatars / photos : lecture publique (URLs storage.googleapis.com).
# GCP interdit une condition allUsers qui exclurait visit-reports/ —
# la protection PHI repose sur l’app (pas d’URL + stream auth + UUID).
# Residual : préfixes sensibles (documents/, afmps-imports/, …) restent
# lisibles si le chemin objet est connu — utiliser des clés opaques
# (surtout AFMPS_IMPORT_OBJECT_KEY, pas latest.csv en prod).
# les objets PHI n’exposent pas d’URL publique (media.Upload → "").
echo "→ IAM objectViewer public (allUsers) pour médias non-PHI"
gcloud storage buckets add-iam-policy-binding "gs://${BUCKET}" \
  --member="allUsers" \
  --role="roles/storage.objectViewer" \
  --quiet >/dev/null 2>&1 || true

# Filet de secours sur les packs de partage PHI (dossier ZIP + consultation PDF) :
# le token expire à 24 h et la purge applicative les efface, mais si les deux échouent
# le bucket nettoie.
# --lifecycle-file remplace TOUTES les règles : on fusionne avec l'existant.
echo "→ Lifecycle : suppression de dossier-shares/** et consultation-shares/** au-delà de 2 jours (merge)"
LIFECYCLE_FILE="$(mktemp)"
trap 'rm -f "$LIFECYCLE_FILE"' EXIT
python3 - "$BUCKET" "$GCP_PROJECT_ID" "$LIFECYCLE_FILE" <<'PY'
import json, subprocess, sys

bucket, project, out = sys.argv[1], sys.argv[2], sys.argv[3]
desired = {
    "action": {"type": "Delete"},
    "condition": {"age": 2, "matchesPrefix": ["dossier-shares/", "consultation-shares/"]},
}

raw = subprocess.check_output(
    [
        "gcloud", "storage", "buckets", "describe", f"gs://{bucket}",
        f"--project={project}", "--format=json",
    ],
    text=True,
)
meta = json.loads(raw)
# gcloud JSON : lifecycle.rule ou lifecycle_config.rule selon la version CLI.
lifecycle = meta.get("lifecycle") or meta.get("lifecycle_config") or {}
rules = list(lifecycle.get("rule") or [])

def is_phi_share_lifecycle_rule(rule: dict) -> bool:
    cond = rule.get("condition") or {}
    prefixes = cond.get("matchesPrefix") or []
    if not isinstance(prefixes, list):
        return False
    # Ancienne règle dossier-only ou règle fusionnée actuelle.
    return set(prefixes) in (
        {"dossier-shares/"},
        {"dossier-shares/", "consultation-shares/"},
    )

kept = [r for r in rules if not is_phi_share_lifecycle_rule(r)]
kept.append(desired)
payload = {"rule": kept}
with open(out, "w", encoding="utf-8") as f:
    json.dump(payload, f)
print(f"  {len(kept)} règle(s) lifecycle (dont dossier-shares/ + consultation-shares/)")
PY
gcloud storage buckets update "gs://${BUCKET}" \
  --project="$GCP_PROJECT_ID" \
  --lifecycle-file="$LIFECYCLE_FILE" \
  --quiet >/dev/null

echo ""
echo "Configurer Cloud Run API :"
echo "  GCS_MEDIA_BUCKET=${BUCKET}"
echo "Done. PHI visit-reports : pas d’URL publique (API Open auth uniquement)."
echo "Avertissement : un objet visit-reports/* reste lisible si son chemin UUID est connu"
echo "  (UBLA + allUsers). Mitigation : clés aléatoires + pas d’URL retournée + purge finalize."
