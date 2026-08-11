<template>
  <div class="pro-app">
    <template v-if="shellReady">
      <ProTopbar home-link="/admin" :show-notifications="false" />
      <div class="pro-app-shell">
        <ProSidebar :items="navItems" />
        <div class="pro-app-body">
          <main class="pro-main main">
            <div class="pro-main-inner">
              <slot />
            </div>
          </main>
        </div>
      </div>
    </template>
    <main v-else class="pro-main main">
      <div class="pro-main-inner">
        <slot />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import type { ProNavItem } from '~/components/pro/ProSidebar.vue'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

const { t } = useI18n()
const { user, fetchUser } = useProUser()
const { isStagingLike } = useAppEnv()
const runtimeConfig = useRuntimeConfig()
const billitOn = computed(() => isPublicFlagOn(runtimeConfig.public.billitEnabled))
const pharmacyOn = computed(() => isPublicFlagOn(runtimeConfig.public.pharmacyEnabled))
const researchOn = computed(() => isPublicFlagOn(runtimeConfig.public.researchEnabled))
if (!user.value) {
  await fetchUser().catch(() => null)
}
const shellReady = computed(() => !!user.value?.role)
const isDevRole = computed(() => user.value?.role === 'dev')

const navItems = computed<ProNavItem[]>(() => {
  const ops = t('nav.section.ops')
  const imports = t('nav.section.imports')
  const content = t('nav.section.content')
  const modules = t('nav.section.modules')
  const salesForce = t('nav.section.salesForce')
  const ai = t('nav.section.ai')
  const billing = t('nav.section.billing')
  const tagDev = t('nav.tagDev')

  const opsItems: ProNavItem[] = [
    { to: '/admin', label: t('nav.adminDashboard'), exact: true, icon: 'admin', section: ops },
    { to: '/admin/users', label: t('nav.adminUsers'), icon: 'users', section: ops },
    { to: '/admin/support', label: t('nav.adminSupport'), icon: 'support_agent', section: ops },
    { to: '/admin/runtime-flags', label: t('nav.adminRuntimeFlags'), icon: 'tune', section: ops },
  ]

  const moduleItems: ProNavItem[] = [
    ...(isPublicFlagOn(runtimeConfig.public.pacsEnabled)
      ? [{ to: '/admin/pacs', label: t('nav.adminPacs'), icon: 'image' as const, section: modules, tag: tagDev }]
      : []),
    ...(researchOn.value
      ? [{ to: '/admin/research', label: t('nav.adminResearch'), icon: 'analytics' as const, section: modules, tag: tagDev }]
      : []),
  ]

  // DEV = support IT léger : tickets, users, flags (pas sales / brand / AI / billing).
  // Pas de /usecases : middleware staging-usecases = admin/commercial/manager seulement.
  if (isDevRole.value) {
    return [...opsItems, ...moduleItems]
  }

  return [
    ...opsItems,
    { to: '/admin/client-imports', label: t('nav.adminClientImports'), icon: 'description', section: imports },
    ...(pharmacyOn.value
      ? [
          { to: '/admin/afmps-imports', label: t('nav.adminAfmps'), icon: 'medication' as const, section: imports, tag: tagDev },
          { to: '/admin/compendium-imports', label: t('nav.adminCompendium'), icon: 'medication' as const, section: imports, tag: tagDev },
        ]
      : []),
    { to: '/admin/brand-assets', label: t('nav.adminBrandAssets'), icon: 'description', section: content },
    productFlowsNavItem(t('nav.productFlows'), content),
    { to: '/nouveautes', label: t('nav.nouveautes'), icon: 'newspaper', section: content },
    presentationNavItem(t('presentation.ui.navLabel'), content),
    ...(isStagingLike.value
      ? [usecasesNavItem(t('nav.usecases'), content)]
      : []),
    ...moduleItems,
    { to: '/admin/commercials', label: t('nav.adminCommercials'), icon: 'users', section: salesForce },
    { to: '/admin/vet-pool', label: t('nav.adminVetPool'), icon: 'pets', section: salesForce },
    { to: '/admin/filiation', label: t('nav.adminFiliation'), icon: 'account_tree', section: salesForce },
    { to: '/admin/sales-branches', label: t('nav.adminSalesBranches'), icon: 'account_tree', section: salesForce },
    { to: '/admin/prospects', label: t('nav.adminProspects'), icon: 'requests', section: salesForce },
    aiFlowsNavItem(t('nav.aiFlows'), ai),
    { to: '/admin/ai-modules', label: t('nav.adminAiModules'), icon: 'record_voice_over', section: ai },
    ...(isPublicFlagOn(runtimeConfig.public.aiCrAdvancedEnabled)
      ? [{ to: '/admin/rag', label: t('nav.adminRag'), icon: 'menu_book' as const, section: ai, tag: tagDev }]
      : []),
    { to: '/admin/training', label: t('nav.adminTraining'), icon: 'record_voice_over', section: ai },
    { to: '/admin/payments', label: t('nav.adminPayments'), icon: 'payments', section: billing },
    ...(billitOn.value
      ? [{ to: '/admin/invoicing', label: t('nav.adminInvoicing'), icon: 'receipt' as const, section: billing }]
      : []),
    { to: '/admin/stripe-catalog', label: t('nav.adminStripeCatalog'), icon: 'payments', section: billing },
    { to: '/admin/commissions', label: t('nav.adminCommissions'), icon: 'payments', section: billing },
    { to: '/admin/commercial-commissions', label: t('nav.adminCommercialCommissions'), icon: 'payments', section: billing },
    { to: '/admin/commercial-bonuses', label: t('nav.adminCommercialBonuses'), icon: 'payments', section: billing },
  ]
})

onMounted(() => {
  if (!user.value) void fetchUser().catch(() => {})
})

</script>
