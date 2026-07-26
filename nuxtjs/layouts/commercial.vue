<template>
  <div class="pro-app">
    <ProTopbar home-link="/commercial" settings-link="/commercial/settings" :show-notifications="false" />
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
  { to: '/commercial/prospects', label: t('nav.commercialProspects'), icon: 'requests', section: t('nav.section.pipeline') },
  { to: '/commercial/vets', label: t('nav.commercialVets'), icon: 'users', section: t('nav.section.pipeline') },
  { to: '/produits', label: t('nav.products'), icon: 'description', section: t('nav.section.offer') },
  { to: '/commercial/pitch', label: t('nav.commercialPitch'), icon: 'campaign', section: t('nav.section.offer') },
  { to: '/commercial/pitch-deck', label: t('pitchDeck.ui.navLabel'), icon: 'slideshow', section: t('nav.section.offer') },
  { to: '/commercial/competition', label: t('nav.commercialCompetition'), icon: 'analytics', section: t('nav.section.offer') },
  { to: '/commercial/training', label: t('nav.commercialTraining'), icon: 'phone_in_talk', section: t('nav.section.ai') },
  { to: '/commercial/ai-modules', label: t('nav.commercialAiModules'), icon: 'record_voice_over', section: t('nav.section.ai') },
  { to: '/commercial/ai-cr-playbook', label: t('nav.commercialAiPlaybook'), icon: 'description', section: t('nav.section.ai') },
  { to: '/commercial/network', label: t('nav.commercialNetwork'), icon: 'account_tree', section: t('nav.section.network') },
  { to: '/commercial/commissions', label: t('nav.commercialCommissions'), icon: 'payments', section: t('nav.section.payout') },
  { to: '/commercial/settings', label: t('nav.commercialSettings'), icon: 'settings', section: t('nav.section.payout') },
])

onMounted(() => { void fetchUser().catch(() => {}) })

</script>
