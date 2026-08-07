<template>
  <div class="pro-app">
    <template v-if="shellReady">
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

const { t } = useI18n()
const { user, fetchUser } = useProUser()
const { isStagingLike } = useAppEnv()
if (!user.value) {
  await fetchUser().catch(() => null)
}
const shellReady = computed(() => !!user.value?.role)

const navItems = computed<ProNavItem[]>(() => [
  { to: '/commercial', label: t('nav.commercialDashboard'), exact: true, icon: 'dashboard', section: t('nav.section.day') },
  { to: '/commercial/agenda', label: t('nav.commercialAgenda'), icon: 'event', section: t('nav.section.day') },
  { to: '/commercial/prospects', label: t('nav.commercialProspects'), icon: 'requests', section: t('nav.section.pipeline') },
  { to: '/commercial/email-templates', label: t('nav.commercialEmailTemplates'), icon: 'mail', section: t('nav.section.pipeline') },
  { to: '/commercial/emails', label: t('nav.commercialEmails'), icon: 'outbox', section: t('nav.section.pipeline') },
  { to: '/commercial/vets', label: t('nav.commercialVets'), icon: 'users', section: t('nav.section.pipeline') },
  { to: '/commercial/filiation', label: t('nav.commercialFiliation'), icon: 'account_tree', section: t('nav.section.network') },
  { to: '/commercial/network', label: t('nav.commercialNetwork'), icon: 'account_tree', section: t('nav.section.network') },
  { to: '/produits', label: t('nav.products'), icon: 'description', section: t('nav.section.offer') },
  { to: '/nouveautes', label: t('nav.nouveautes'), icon: 'newspaper', section: t('nav.section.offer') },
  productFlowsNavItem(t('nav.productFlows'), t('nav.section.offer')),
  presentationNavItem(t('presentation.ui.navLabel'), t('nav.section.offer')),
  { to: '/commercial/pitch', label: t('nav.commercialPitch'), icon: 'campaign', section: t('nav.section.offer') },
  { to: '/commercial/asv-memo', label: t('nav.commercialAsvMemo'), icon: 'support_agent', section: t('nav.section.offer') },
  { to: '/commercial/brochure', label: t('nav.commercialBrochure'), icon: 'picture_as_pdf', section: t('nav.section.offer') },
  { to: '/commercial/competition', label: t('nav.commercialCompetition'), icon: 'analytics', section: t('nav.section.offer') },
  ...(isStagingLike.value
    ? [usecasesNavItem(t('nav.usecases'), t('nav.section.offer'))]
    : []),
  { to: '/commercial/training', label: t('nav.commercialTraining'), icon: 'phone_in_talk', section: t('nav.section.ai') },
  aiFlowsNavItem(t('nav.aiFlows'), t('nav.section.ai')),
  { to: '/commercial/ai-modules', label: t('nav.commercialAiModules'), icon: 'record_voice_over', section: t('nav.section.ai') },
  { to: '/commercial/ai-cr-playbook', label: t('nav.commercialAiPlaybook'), icon: 'description', section: t('nav.section.ai') },
  { to: '/commercial/commissions', label: t('nav.commercialCommissions'), icon: 'payments', section: t('nav.section.payout') },
  { to: '/commercial/settings', label: t('nav.commercialSettings'), icon: 'settings', section: t('nav.section.account') },
])

onMounted(() => {
  if (!user.value) void fetchUser().catch(() => {})
})

</script>
