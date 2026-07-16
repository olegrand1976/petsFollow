<template>
  <div data-testid="commercial-dashboard-page">
    <ProPageHeader
      :title="$t('commercial.dashboard.title')"
      :subtitle="$t('commercial.dashboard.subtitle')"
    />
    <div v-if="overview" class="pro-grid-kpi">
      <ProKpi :value="overview.assignedVets" :label="$t('commercial.dashboard.assignedVets')" />
      <ProKpi :value="overview.prospectsTotal" :label="$t('commercial.dashboard.prospects')" />
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

onMounted(async () => {
  const res: any = await $fetch('/api/commercial/overview')
  overview.value = res.data ?? res
})
</script>
