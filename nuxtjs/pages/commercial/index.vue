<template>
  <div data-testid="commercial-dashboard-page">
    <ProPageHeader
      :title="$t('commercial.dashboard.title')"
      :subtitle="$t('commercial.dashboard.subtitle')"
    >
      <template #actions>
        <ProButton
          variant="secondary"
          test-id="commercial-app-invite-open"
          @click="appInviteOpen = true"
        >
          <ProIcon name="qr_code_2" />
          {{ $t('commercial.appInvite.open') }}
        </ProButton>
      </template>
    </ProPageHeader>
    <ProAppInviteModal v-model:open="appInviteOpen" :title="$t('commercial.appInvite.title')" />
    <div v-if="overview" class="pro-grid-kpi">
      <ProKpi :value="overview.assignedVets" :label="$t('commercial.dashboard.assignedVets')" />
      <ProKpi :value="overview.prospectsTotal" :label="$t('commercial.dashboard.prospects')" />
      <ProKpi :value="overview.directoryProspects ?? 0" :label="$t('commercial.dashboard.directory')" />
      <ProKpi :value="overview.prospectsNew" :label="$t('commercial.dashboard.prospectsNew')" />
      <ProKpi :value="overview.prospectsConverted" :label="$t('commercial.dashboard.converted')" />
      <ProKpi :value="formatCurrency(overview.monthEarnedCents)" :label="$t('commercial.dashboard.monthEarned')" />
      <ProKpi :value="formatCurrency(overview.lifetimeEarnedCents)" :label="$t('commercial.dashboard.lifetimeEarned')" />
      <ProKpi :value="formatCurrency(overview.linkedSubscriptionRevenueCents)" :label="$t('commercial.dashboard.subRevenue')" />
      <ProKpi :value="formatCurrency(overview.linkedAddonRevenueCents)" :label="$t('commercial.dashboard.addonRevenue')" />
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { formatCurrency } = useFormatters()
const overview = ref<any>(null)
const appInviteOpen = ref(false)

onMounted(async () => {
  const res: any = await $fetch('/api/commercial/overview')
  overview.value = res.data ?? res
})
</script>
