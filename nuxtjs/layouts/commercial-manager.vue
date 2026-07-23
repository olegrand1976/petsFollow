<template>
  <div class="pro-app">
    <ProTopbar home-link="/commercial-manager" :show-notifications="false" />
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
  { to: '/commercial-manager', label: t('nav.managerDashboard'), exact: true, icon: 'dashboard' },
  { to: '/commercial-manager/suivi', label: t('nav.managerFollowups'), icon: 'event' },
  { to: '/commercial-manager/prospects', label: t('nav.managerProspects'), icon: 'requests' },
  { to: '/commercial', label: t('nav.managerPortfolio'), icon: 'users' },
  { to: '/commercial/pitch', label: t('nav.commercialPitch'), icon: 'campaign' },
  { to: '/commercial-manager/training', label: t('nav.commercialTraining'), icon: 'phone_in_talk' },
  { to: '/produits', label: t('nav.products'), icon: 'description' },
])

onMounted(() => { void fetchUser().catch(() => {}) })

</script>
