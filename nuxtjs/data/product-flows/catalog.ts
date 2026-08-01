import type { FlowLink, FlowProfileId, ProductFlowsCatalog } from './types'

/**
 * Carte des flux petsFollow par profil + liens croisés.
 * Libellés via i18n `productFlows.*` (pas de texte hardcodé ici).
 */
export const productFlowsCatalog: ProductFlowsCatalog = {
  profiles: [
    {
      id: 'vet',
      surface: 'web_pro',
      icon: 'medical_services',
      featureIds: [
        'onboarding',
        'dashboard',
        'messaging',
        'hr_vitals',
        'link_requests',
        'dossier_share',
        'commissions',
        'ai_cr',
        'practice_desk',
      ],
      stepIds: ['vet_register', 'vet_onboarding', 'vet_dashboard', 'vet_care', 'vet_earn'],
    },
    {
      id: 'client',
      surface: 'flutter_client',
      icon: 'person',
      featureIds: [
        'onboarding',
        'billing_client',
        'hr_vitals',
        'messaging',
        'care_horse',
        'dossier_share',
        'link_requests',
      ],
      stepIds: [
        'client_register',
        'client_pet',
        'client_pay',
        'client_active',
        'client_engage',
      ],
    },
    {
      id: 'care_pro',
      surface: 'flutter_pro_light',
      icon: 'handshake',
      featureIds: ['pro_light_agenda', 'dossier_share', 'messaging'],
      stepIds: ['care_login', 'care_agenda', 'care_notes', 'care_sync'],
    },
    {
      id: 'commercial',
      surface: 'web_pro',
      icon: 'campaign',
      featureIds: ['prospects', 'encode', 'network', 'commissions', 'ai_cr'],
      stepIds: [
        'co_login',
        'co_overview',
        'co_prospects',
        'co_encode',
        'co_activate',
        'co_earn',
      ],
    },
    {
      id: 'commercial_manager',
      surface: 'web_pro',
      icon: 'groups',
      featureIds: ['team_kpi', 'prospects', 'encode', 'network', 'commissions'],
      stepIds: ['cm_login', 'cm_team', 'cm_followups', 'cm_self', 'cm_encode'],
    },
    {
      id: 'admin',
      surface: 'web_pro',
      icon: 'admin_panel_settings',
      featureIds: ['ops_users', 'ops_billing', 'network', 'ai_cr', 'research_obs'],
      stepIds: ['ad_login', 'ad_metrics', 'ad_users', 'ad_sales', 'ad_close'],
    },
    {
      id: 'practice_staff',
      surface: 'web_pro',
      icon: 'support_agent',
      featureIds: ['practice_desk', 'messaging', 'dashboard'],
      stepIds: ['eq_login', 'eq_desk', 'eq_clients', 'eq_msg'],
    },
    {
      id: 'research',
      surface: 'web_pro',
      icon: 'analytics',
      tagDev: true,
      featureIds: ['research_obs'],
      stepIds: ['re_optin', 're_etl', 're_login', 're_obs'],
    },
  ],
  links: [
    {
      id: 'msg_vet_client',
      fromProfile: 'vet',
      toProfile: 'client',
      featureId: 'messaging',
    },
    {
      id: 'link_req_client_vet',
      fromProfile: 'client',
      toProfile: 'vet',
      featureId: 'link_requests',
    },
    {
      id: 'hr_client_vet',
      fromProfile: 'client',
      toProfile: 'vet',
      featureId: 'hr_vitals',
    },
    {
      id: 'pay_client_commissions',
      fromProfile: 'client',
      toProfile: 'commercial',
      featureId: 'commissions',
    },
    {
      id: 'pay_client_vet_commissions',
      fromProfile: 'client',
      toProfile: 'vet',
      featureId: 'commissions',
    },
    {
      id: 'encode_co_vet',
      fromProfile: 'commercial',
      toProfile: 'vet',
      featureId: 'encode',
    },
    {
      id: 'encode_co_client',
      fromProfile: 'commercial',
      toProfile: 'client',
      featureId: 'encode',
    },
    {
      id: 'care_notes_vet',
      fromProfile: 'care_pro',
      toProfile: 'vet',
      featureId: 'pro_light_agenda',
    },
    {
      id: 'dossier_client_pro',
      fromProfile: 'client',
      toProfile: 'vet',
      featureId: 'dossier_share',
    },
    {
      id: 'manager_team_co',
      fromProfile: 'commercial_manager',
      toProfile: 'commercial',
      featureId: 'team_kpi',
    },
    {
      id: 'admin_sales_co',
      fromProfile: 'admin',
      toProfile: 'commercial',
      featureId: 'network',
    },
    {
      id: 'staff_vet_desk',
      fromProfile: 'practice_staff',
      toProfile: 'vet',
      featureId: 'practice_desk',
    },
    {
      id: 'research_vet_optin',
      fromProfile: 'vet',
      toProfile: 'research',
      featureId: 'research_obs',
    },
  ],
}

/** Liens croisés touchant un profil (from ou to). */
export function linksForProfile(id: FlowProfileId): FlowLink[] {
  return productFlowsCatalog.links.filter(
    (l) => l.fromProfile === id || l.toProfile === id,
  )
}

const profileIdSet = new Set<FlowProfileId>(
  productFlowsCatalog.profiles.map((p) => p.id),
)

/** Valide un id de profil (ex. query `?profile=` ; accepte aussi un tableau Nuxt). */
export function parseFlowProfileId(raw: unknown): FlowProfileId | null {
  const value = Array.isArray(raw) ? raw[0] : raw
  if (typeof value !== 'string' || !value) return null
  return profileIdSet.has(value as FlowProfileId) ? (value as FlowProfileId) : null
}
