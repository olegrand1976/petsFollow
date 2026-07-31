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
if (!user.value) {
  await fetchUser().catch(() => null)
}
const shellReady = computed(() => !!user.value?.role)
const isDevRole = computed(() => user.value?.role === 'dev')

const navItems = computed<ProNavItem[]>(() => {
  // DEV = support IT léger : tickets, users, flags (pas sales / brand / AI / billing).
  if (isDevRole.value) {
    // Pas de /usecases : middleware staging-usecases = admin/commercial/manager seulement.
    return [
      { to: '/admin', label: t('nav.adminDashboard'), exact: true, icon: 'admin', section: t('nav.section.ops') },
      { to: '/admin/users', label: t('nav.adminUsers'), icon: 'users', section: t('nav.section.ops') },
      { to: '/admin/support', label: t('nav.adminSupport'), icon: 'support_agent', section: t('nav.section.ops') },
      { to: '/admin/runtime-flags', label: t('nav.adminRuntimeFlags'), icon: 'tune', section: t('nav.section.ops') },
      ...(isPublicFlagOn(runtimeConfig.public.pacsEnabled)
        ? [{ to: '/admin/pacs', label: t('nav.adminPacs'), icon: 'image', section: t('nav.section.ops'), tag: t('nav.tagDev') }]
        : []),
    ]
  }
  return [
    { to: '/admin', label: t('nav.adminDashboard'), exact: true, icon: 'admin', section: t('nav.section.ops') },
    { to: '/admin/users', label: t('nav.adminUsers'), icon: 'users', section: t('nav.section.ops') },
    { to: '/admin/client-imports', label: t('nav.adminClientImports'), icon: 'description', section: t('nav.section.ops') },
    { to: '/admin/compendium-imports', label: t('nav.adminCompendium'), icon: 'medication', section: t('nav.section.ops'), tag: t('nav.tagDev') },
    { to: '/admin/brand-assets', label: t('nav.adminBrandAssets'), icon: 'description', section: t('nav.section.ops') },
    { to: '/admin/support', label: t('nav.adminSupport'), icon: 'support_agent', section: t('nav.section.ops') },
    { to: '/admin/runtime-flags', label: t('nav.adminRuntimeFlags'), icon: 'tune', section: t('nav.section.ops') },
    ...(isPublicFlagOn(runtimeConfig.public.pacsEnabled)
      ? [{ to: '/admin/pacs', label: t('nav.adminPacs'), icon: 'image', section: t('nav.section.ops'), tag: t('nav.tagDev') }]
      : []),
    ...(isStagingLike.value
      ? [usecasesNavItem(t('nav.usecases'), t('nav.section.ops'))]
      : []),
    { to: '/admin/commercials', label: t('nav.adminCommercials'), icon: 'users', section: t('nav.section.salesForce') },
    { to: '/admin/vet-pool', label: t('nav.adminVetPool'), icon: 'pets', section: t('nav.section.salesForce') },
    { to: '/admin/filiation', label: t('nav.adminFiliation'), icon: 'account_tree', section: t('nav.section.salesForce') },
    { to: '/admin/sales-branches', label: t('nav.adminSalesBranches'), icon: 'account_tree', section: t('nav.section.salesForce') },
    { to: '/admin/prospects', label: t('nav.adminProspects'), icon: 'requests', section: t('nav.section.salesForce') },
    { to: '/admin/ai-modules', label: t('nav.adminAiModules'), icon: 'record_voice_over', section: t('nav.section.ai') },
    { to: '/admin/training', label: t('nav.adminTraining'), icon: 'record_voice_over', section: t('nav.section.ai') },
    { to: '/admin/payments', label: t('nav.adminPayments'), icon: 'payments', section: t('nav.section.billing') },
    ...(billitOn.value
      ? [{ to: '/admin/invoicing', label: t('nav.adminInvoicing'), icon: 'receipt', section: t('nav.section.billing') }]
      : []),
    { to: '/admin/stripe-catalog', label: t('nav.adminStripeCatalog'), icon: 'payments', section: t('nav.section.billing') },
    { to: '/admin/commissions', label: t('nav.adminCommissions'), icon: 'payments', section: t('nav.section.billing') },
    { to: '/admin/commercial-commissions', label: t('nav.adminCommercialCommissions'), icon: 'payments', section: t('nav.section.billing') },
    { to: '/admin/commercial-bonuses', label: t('nav.adminCommercialBonuses'), icon: 'payments', section: t('nav.section.billing') },
  ]
})

onMounted(() => {
  if (!user.value) void fetchUser().catch(() => {})
})

</script>
