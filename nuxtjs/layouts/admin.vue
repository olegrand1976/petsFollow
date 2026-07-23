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
  { to: '/admin', label: t('nav.adminDashboard'), exact: true, icon: 'admin' },
  { to: '/admin/users', label: t('nav.adminUsers'), icon: 'users' },
  { to: '/admin/client-imports', label: t('nav.adminClientImports'), icon: 'description' },
  { to: '/admin/commercials', label: t('nav.adminCommercials'), icon: 'users' },
  { to: '/admin/prospects', label: t('nav.adminProspects'), icon: 'requests' },
  { to: '/admin/payments', label: t('nav.adminPayments'), icon: 'payments' },
  { to: '/admin/commissions', label: t('nav.adminCommissions'), icon: 'payments' },
  { to: '/admin/commercial-commissions', label: t('nav.adminCommercialCommissions'), icon: 'payments' },
  { to: '/admin/commercial-bonuses', label: t('nav.adminCommercialBonuses'), icon: 'payments' },
  { to: '/admin/training', label: t('nav.adminTraining'), icon: 'record_voice_over' },
])

onMounted(() => { void fetchUser().catch(() => {}) })

</script>
