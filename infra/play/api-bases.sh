# Canonical public API / site URLs for Play AAB and GCP deploy docs.
# Sourced by infra/play/build-play-bundle.sh (and Make targets).
#
# Staging  = Cloud Run actuel (seedable) — Play Internal testing only.
# Production = domaine petsfollow.app — activé avec deploy GitHub `main` + GCP prod.

PETSFOLLOW_API_STAGING="${PETSFOLLOW_API_STAGING:-https://api.petsfollow.ll-it-sc.be}"
PETSFOLLOW_SITE_STAGING="${PETSFOLLOW_SITE_STAGING:-https://petsfollow.ll-it-sc.be}"

# Cibles prod (DNS / LB / Cloud Run à provisionner avant le 1er AAB Production).
PETSFOLLOW_API_PROD="${PETSFOLLOW_API_PROD:-https://api.petsfollow.app}"
PETSFOLLOW_SITE_PROD="${PETSFOLLOW_SITE_PROD:-https://petsfollow.app}"
