#!/bin/sh
# Render Orthanc config from env, then exec Orthanc.
# Cloud Run sets PORT (default 8080). Secrets via env / Secret Manager mounts.
set -eu

TEMPLATE="${ORTHANC_TEMPLATE:-/etc/orthanc/orthanc.json.template}"
OUT="${ORTHANC_CONFIG:-/tmp/orthanc.json}"

ORTHANC_USER="${ORTHANC_USER:-petsfollow}"
ORTHANC_PASSWORD="${ORTHANC_PASSWORD:-changeme}"
# Cloud Run IAM gates access (ingress=internal) → Orthanc auth off by default.
# Local / non-IAM : ORTHANC_AUTH_ENABLED=true
AUTH_ENABLED="${ORTHANC_AUTH_ENABLED:-false}"
GCS_DICOM_BUCKET="${GCS_DICOM_BUCKET:-petsfollow-dicom}"
GCS_SA_FILE="${GCS_SA_FILE:-}"
PG_HOST="${PG_HOST:?PG_HOST required}"
PG_PORT="${PG_PORT:-5432}"
PG_DATABASE="${PG_DATABASE:-orthanc}"
PG_USERNAME="${PG_USERNAME:-orthanc}"
PG_PASSWORD="${PG_PASSWORD:?PG_PASSWORD required}"
HTTP_PORT="${PORT:-8080}"
case "$AUTH_ENABLED" in
  true|1|TRUE|True) AUTH_JSON=true ;;
  *) AUTH_JSON=false ;;
esac

# Escape sed replacement metacharacters in secrets.
esc() {
  printf '%s' "$1" | sed -e 's/[\\/&|]/\\&/g'
}

sed \
  -e "s|__ORTHANC_USER__|$(esc "$ORTHANC_USER")|g" \
  -e "s|__ORTHANC_PASSWORD__|$(esc "$ORTHANC_PASSWORD")|g" \
  -e "s|__AUTH_ENABLED__|${AUTH_JSON}|g" \
  -e "s|__GCS_DICOM_BUCKET__|$(esc "$GCS_DICOM_BUCKET")|g" \
  -e "s|__GCS_SA_FILE__|$(esc "$GCS_SA_FILE")|g" \
  -e "s|__PG_HOST__|$(esc "$PG_HOST")|g" \
  -e "s|__PG_PORT__|${PG_PORT}|g" \
  -e "s|__PG_DATABASE__|$(esc "$PG_DATABASE")|g" \
  -e "s|__PG_USERNAME__|$(esc "$PG_USERNAME")|g" \
  -e "s|__PG_PASSWORD__|$(esc "$PG_PASSWORD")|g" \
  -e "s|\"HttpPort\": 8080|\"HttpPort\": ${HTTP_PORT}|g" \
  "$TEMPLATE" >"$OUT"

exec Orthanc "$OUT" --verbose
