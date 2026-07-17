<template>
  <div class="pro-app">
    <ProTopbar home-link="/commercial" :show-notifications="false" />
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
  { to: '/commercial', label: t('nav.commercialDashboard'), exact: true, icon: 'dashboard' },
  { to: '/commercial/vets', label: t('nav.commercialVets'), icon: 'users' },
  { to: '/commercial/prospects', label: t('nav.commercialProspects'), icon: 'requests' },
  { to: '/commercial/commissions', label: t('nav.commercialCommissions'), icon: 'payments' },
  { to: '/commercial/settings', label: t('nav.commercialSettings'), icon: 'settings' },
  { to: '/commercial/pitch', label: t('nav.commercialPitch'), icon: 'settings' },
  { to: '/produits', label: t('nav.products'), icon: 'description' },
])

onMounted(() => fetchUser())
</script>
