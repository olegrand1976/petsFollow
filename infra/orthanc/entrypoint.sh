#!/bin/sh
# Deprecated — Cloud Run Orthanc must use the Orthanc Team docker-entrypoint.sh
# so plugins (PostgreSQL / GoogleCloudStorage / DicomWeb) are symlinked from
# plugins-available before Orthanc starts. See deploy/Dockerfile.orthanc.
#
# Kept as a no-op stub so any lingering docs/scripts that reference this path
# fail loudly instead of silently running SQLite+/tmp.
echo "FATAL: infra/orthanc/entrypoint.sh must not be used as Cloud Run ENTRYPOINT." >&2
echo "Use the Orthanc Team image entrypoint + ORTHANC__* / *_PLUGIN_ENABLED env." >&2
exit 1
