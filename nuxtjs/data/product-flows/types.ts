/** Profils métier présentés dans /flux (orientation commerciale / admin). */
export type FlowProfileId =
  | 'vet'
  | 'client'
  | 'care_pro'
  | 'commercial'
  | 'commercial_manager'
  | 'admin'
  | 'practice_staff'
  | 'research'

export type FlowSurface = 'web_pro' | 'flutter_client' | 'flutter_pro_light'

export type FlowFeatureId =
  | 'onboarding'
  | 'dashboard'
  | 'messaging'
  | 'hr_vitals'
  | 'link_requests'
  | 'dossier_share'
  | 'billing_client'
  | 'care_horse'
  | 'commissions'
  | 'prospects'
  | 'encode'
  | 'network'
  | 'team_kpi'
  | 'ops_users'
  | 'ops_billing'
  | 'ai_cr'
  | 'pro_light_agenda'
  | 'practice_desk'
  | 'research_obs'

export type FlowStepId =
  | 'vet_register'
  | 'vet_onboarding'
  | 'vet_dashboard'
  | 'vet_care'
  | 'vet_earn'
  | 'client_register'
  | 'client_pet'
  | 'client_pay'
  | 'client_active'
  | 'client_engage'
  | 'care_login'
  | 'care_agenda'
  | 'care_notes'
  | 'care_sync'
  | 'co_login'
  | 'co_overview'
  | 'co_prospects'
  | 'co_encode'
  | 'co_activate'
  | 'co_earn'
  | 'cm_login'
  | 'cm_team'
  | 'cm_followups'
  | 'cm_self'
  | 'cm_encode'
  | 'ad_login'
  | 'ad_metrics'
  | 'ad_users'
  | 'ad_sales'
  | 'ad_close'
  | 'eq_login'
  | 'eq_desk'
  | 'eq_clients'
  | 'eq_msg'
  | 're_optin'
  | 're_etl'
  | 're_login'
  | 're_obs'

export type FlowLinkId =
  | 'msg_vet_client'
  | 'link_req_client_vet'
  | 'hr_client_vet'
  | 'pay_client_commissions'
  | 'pay_client_vet_commissions'
  | 'encode_co_vet'
  | 'encode_co_client'
  | 'care_notes_vet'
  | 'dossier_client_pro'
  | 'manager_team_co'
  | 'admin_sales_co'
  | 'staff_vet_desk'
  | 'research_vet_optin'

export type FlowProfile = {
  id: FlowProfileId
  surface: FlowSurface
  /** Material Symbols Outlined ligature. */
  icon: string
  featureIds: FlowFeatureId[]
  stepIds: FlowStepId[]
  tagDev?: boolean
}

export type FlowLink = {
  id: FlowLinkId
  fromProfile: FlowProfileId
  toProfile: FlowProfileId
  featureId: FlowFeatureId
}

export type ProductFlowsCatalog = {
  profiles: FlowProfile[]
  links: FlowLink[]
}
