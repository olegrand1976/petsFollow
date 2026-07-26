<template>
  <div class="pro-app">
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
  </div>
</template>

<script setup lang="ts">
import type { ProNavItem } from '~/components/pro/ProSidebar.vue'

const { t } = useI18n()
const { fetchUser } = useProUser()

const navItems = computed<ProNavItem[]>(() => [
  { to: '/admin', label: t('nav.adminDashboard'), exact: true, icon: 'admin', section: t('nav.section.ops') },
  { to: '/admin/users', label: t('nav.adminUsers'), icon: 'users', section: t('nav.section.ops') },
  { to: '/admin/client-imports', label: t('nav.adminClientImports'), icon: 'description', section: t('nav.section.ops') },
  { to: '/admin/brand-assets', label: t('nav.adminBrandAssets'), icon: 'description', section: t('nav.section.ops') },
  { to: '/admin/commercials', label: t('nav.adminCommercials'), icon: 'users', section: t('nav.section.salesForce') },
  { to: '/admin/sales-branches', label: t('nav.adminSalesBranches'), icon: 'account_tree', section: t('nav.section.salesForce') },
  { to: '/admin/prospects', label: t('nav.adminProspects'), icon: 'requests', section: t('nav.section.salesForce') },
  { to: '/admin/ai-modules', label: t('nav.adminAiModules'), icon: 'record_voice_over', section: t('nav.section.salesForce') },
  { to: '/admin/training', label: t('nav.adminTraining'), icon: 'record_voice_over', section: t('nav.section.salesForce') },
  { to: '/admin/payments', label: t('nav.adminPayments'), icon: 'payments', section: t('nav.section.billing') },
  { to: '/admin/stripe-catalog', label: t('nav.adminStripeCatalog'), icon: 'payments', section: t('nav.section.billing') },
  { to: '/admin/commissions', label: t('nav.adminCommissions'), icon: 'payments', section: t('nav.section.billing') },
  { to: '/admin/commercial-commissions', label: t('nav.adminCommercialCommissions'), icon: 'payments', section: t('nav.section.billing') },
  { to: '/admin/commercial-bonuses', label: t('nav.adminCommercialBonuses'), icon: 'payments', section: t('nav.section.billing') },
])

onMounted(() => { void fetchUser().catch(() => {}) })

</script>
