#!/usr/bin/env bash
# Domaines custom petsFollow via Load Balancer global + Serverless NEG :
#   petsfollow.ll-it-sc.be     → petsfollow-nuxtjs
#   api.petsfollow.ll-it-sc.be → petsfollow-api
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gcp-env.sh
source "${SCRIPT_DIR}/lib/gcp-env.sh"

gcloud config set project "$GCP_PROJECT_ID" >/dev/null

proxy_cert_names() {
  gcloud compute target-https-proxies describe "$HTTPS_PROXY" \
    --global --project="$GCP_PROJECT_ID" \
    --format='json(sslCertificates)' \
    | python3 -c "import json,sys; data=json.load(sys.stdin); certs=data.get('sslCertificates') or []; print(','.join(u.rsplit('/',1)[-1] for u in certs))"
}

renew_managed_cert() {
  local cert_name="$1"
  local domain="$2"
  local domain_status cert_status existing

  cert_status="$(gcloud compute ssl-certificates describe "$cert_name" \
    --global --project="$GCP_PROJECT_ID" \
    --format='value(managed.status)' 2>/dev/null || true)"
  domain_status="$(gcloud compute ssl-certificates describe "$cert_name" \
    --global --project="$GCP_PROJECT_ID" \
    --format="value(managed.domainStatus['${domain}'])" 2>/dev/null || true)"

  if [[ "$cert_status" == "ACTIVE" && "$domain_status" == "ACTIVE" ]]; then
    echo "  Certificat ${cert_name} déjà ACTIVE"
    return 0
  fi
  if [[ "$domain_status" != "FAILED_NOT_VISIBLE" && "$domain_status" != "PROVISIONING_FAILED" && "$cert_status" != "PROVISIONING_FAILED" ]]; then
    echo "  Certificat ${cert_name} : ${cert_status} / ${domain}=${domain_status:-pending}"
    return 0
  fi

  echo "→ Renouvellement certificat ${cert_name} (${domain_status:-${cert_status}}) — DNS doit pointer vers ${LB_IP}"
  existing="$(proxy_cert_names | tr ',' '\n' | grep -vx "$cert_name" | paste -sd, -)"
  gcloud compute target-https-proxies update "$HTTPS_PROXY" \
    --global --project="$GCP_PROJECT_ID" \
    --ssl-certificates="$existing" --quiet
  gcloud compute ssl-certificates delete "$cert_name" \
    --global --project="$GCP_PROJECT_ID" --quiet
  gcloud compute ssl-certificates create "$cert_name" \
    --domains="$domain" --global --project="$GCP_PROJECT_ID"
  if [[ -n "$existing" ]]; then
    gcloud compute target-https-proxies update "$HTTPS_PROXY" \
      --global --project="$GCP_PROJECT_ID" \
      --ssl-certificates="${existing},${cert_name}" --quiet
  else
    gcloud compute target-https-proxies update "$HTTPS_PROXY" \
      --global --project="$GCP_PROJECT_ID" \
      --ssl-certificates="${cert_name}" --quiet
  fi
}

ensure_host_route() {
  local service="$1"
  local domain="$2"
  local neg_name="$3"
  local backend_name="$4"
  local path_matcher="$5"
  local cert_name="$6"

  echo "=== ${domain} → ${service} ==="

  echo "→ NEG serverless (${neg_name})"
  if ! gcloud compute network-endpoint-groups describe "$neg_name" \
    --region="$GCP_RUN_REGION" --project="$GCP_PROJECT_ID" &>/dev/null; then
    gcloud compute network-endpoint-groups create "$neg_name" \
      --region="$GCP_RUN_REGION" \
      --network-endpoint-type=serverless \
      --cloud-run-service="$service" \
      --project="$GCP_PROJECT_ID"
  fi

  echo "→ Backend service (${backend_name})"
  if ! gcloud compute backend-services describe "$backend_name" \
    --global --project="$GCP_PROJECT_ID" &>/dev/null; then
    gcloud compute backend-services create "$backend_name" \
      --global \
      --load-balancing-scheme=EXTERNAL \
      --project="$GCP_PROJECT_ID"
    gcloud compute backend-services add-backend "$backend_name" \
      --global \
      --network-endpoint-group="$neg_name" \
      --network-endpoint-group-region="$GCP_RUN_REGION" \
      --project="$GCP_PROJECT_ID"
  fi

  echo "→ Règle hôte sur ${URL_MAP}"
  if gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
    --format='json(hostRules)' | grep -q '"\*"'; then
    echo "→ Suppression règle hôte wildcard erronée (*)"
    gcloud compute url-maps remove-host-rule "$URL_MAP" \
      --global --project="$GCP_PROJECT_ID" --host='*' 2>/dev/null \
      || echo "  (wildcard * absent ou déjà retiré)"
  fi
  if ! gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
    --format='value(pathMatchers.name)' | tr ';' '\n' | grep -qx "$path_matcher"; then
    gcloud compute url-maps add-path-matcher "$URL_MAP" \
      --global \
      --path-matcher-name="$path_matcher" \
      --default-service="$backend_name" \
      --project="$GCP_PROJECT_ID"
  fi
  if ! gcloud compute url-maps describe "$URL_MAP" --global --project="$GCP_PROJECT_ID" \
    --format='json(hostRules)' | grep -q "\"${domain}\""; then
    gcloud compute url-maps add-host-rule "$URL_MAP" \
      --global \
      --hosts="$domain" \
      --path-matcher-name="$path_matcher" \
      --project="$GCP_PROJECT_ID"
  else
    echo "  Règle hôte ${domain} déjà présente"
  fi

  echo "→ Certificat managé (${cert_name})"
  if ! gcloud compute ssl-certificates describe "$cert_name" \
    --global --project="$GCP_PROJECT_ID" &>/dev/null; then
    # Un seul cert pour les 2 hôtes (quota SSL_CERTIFICATES = 10)
    if [[ "$cert_name" == "$FRONTEND_CERT_NAME" && "$FRONTEND_CERT_NAME" == "$API_CERT_NAME" ]]; then
      gcloud compute ssl-certificates create "$cert_name" \
        --domains="${CUSTOM_DOMAIN},${API_CUSTOM_DOMAIN}" \
        --global \
        --project="$GCP_PROJECT_ID"
    else
      gcloud compute ssl-certificates create "$cert_name" \
        --domains="$domain" \
        --global \
        --project="$GCP_PROJECT_ID"
    fi
  else
    renew_managed_cert "$cert_name" "$domain"
  fi

  EXISTING_CERTS="$(proxy_cert_names)"
  if ! echo ",${EXISTING_CERTS}," | grep -q ",${cert_name},"; then
    gcloud compute target-https-proxies update "$HTTPS_PROXY" \
      --global \
      --ssl-certificates="${EXISTING_CERTS},${cert_name}" \
      --project="$GCP_PROJECT_ID"
  fi

  echo "→ Accès public Cloud Run (${service})"
  gcloud run services add-iam-policy-binding "$service" \
    --region="$GCP_RUN_REGION" \
    --member="allUsers" \
    --role="roles/run.invoker" \
    --project="$GCP_PROJECT_ID" \
    --quiet 2>/dev/null || true
}

ensure_host_route \
  "$FRONTEND_SERVICE" \
  "$CUSTOM_DOMAIN" \
  "$FRONTEND_NEG_NAME" \
  "$FRONTEND_BACKEND_NAME" \
  "$FRONTEND_PATH_MATCHER" \
  "$FRONTEND_CERT_NAME"

ensure_host_route \
  "$API_SERVICE" \
  "$API_CUSTOM_DOMAIN" \
  "$API_NEG_NAME" \
  "$API_BACKEND_NAME" \
  "$API_PATH_MATCHER" \
  "$API_CERT_NAME"

cat <<EOF

OK — prochaines étapes manuelles (OVH, zone ll-it-sc.be) :

1. Créer deux enregistrements A :
   - Sous-domaine : petsfollow
     Cible        : ${LB_IP}
   - Sous-domaine : api.petsfollow
     Cible        : ${LB_IP}
2. Attendre propagation DNS + certificat ACTIVE :
   gcloud compute ssl-certificates describe ${FRONTEND_CERT_NAME} --global --format='yaml(managed)'
3. Tester :
   curl -I https://${CUSTOM_DOMAIN}/
   curl -I https://${API_CUSTOM_DOMAIN}/health

EOF
