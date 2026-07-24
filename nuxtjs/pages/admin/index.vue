<template>
  <div data-testid="admin-dashboard-page">
    <ProPageHeader
      :title="$t('admin.dashboard.title')"
      :subtitle="$t('admin.dashboard.subtitle')"
    />
    <div v-if="metrics" class="pro-grid-kpi">
      <ProKpi :value="formatCurrency(metrics.totalRevenueCents)" :label="$t('admin.dashboard.totalRevenue')" />
      <ProKpi :value="formatCurrency(metrics.periodRevenueCents)" :label="$t('admin.dashboard.periodRevenue')" />
      <ProKpi :value="formatCurrency(metrics.mrrCents)" :label="$t('admin.dashboard.mrr')" />
      <ProKpi :value="metrics.userCount" :label="$t('admin.dashboard.registrations')" />
      <ProKpi :value="metrics.petCount" :label="$t('admin.dashboard.pets')" />
      <ProKpi
        :value="`${metrics.conversionRatePercent.toFixed(1)}%`"
        :label="$t('admin.dashboard.conversion')"
      />
      <ProKpi :value="metrics.pendingPayments" :label="$t('admin.dashboard.pendingPayments')" />
      <ProKpi :value="metrics.pastDueCount" :label="$t('admin.dashboard.pastDue')" />
      <ProKpi :value="metrics.commercialCount ?? 0" :label="$t('admin.dashboard.commercials')" />
      <ProKpi :value="metrics.prospectCount ?? 0" :label="$t('admin.dashboard.prospects')" />
      <ProKpi :value="formatCurrency(metrics.addonRevenueCents ?? 0)" :label="$t('admin.dashboard.addonRevenue')" />
      <ProKpi :value="formatCurrency(metrics.commercialCommissionDueCents ?? 0)" :label="$t('admin.dashboard.commercialDue')" />
    </div>
    <div v-if="metrics" class="pro-grid-2 pro-mt-lg">
      <ProCard :title="$t('admin.dashboard.planBreakdown')">
        <div class="pro-bar-chart">
          <div
            v-for="(count, plan) in metrics.planBreakdown"
            :key="plan"
            class="pro-bar-row"
          >
            <span>{{ planLabel(String(plan)) }}</span>
            <div class="pro-bar-track">
              <div
                class="pro-bar-fill"
                :style="{ width: barWidth(count, planMax) }"
              />
            </div>
            <span>{{ count }}</span>
          </div>
        </div>
      </ProCard>
      <ProCard :title="$t('admin.dashboard.modeBreakdown')">
        <div class="pro-bar-chart">
          <div
            v-for="(count, mode) in metrics.modeBreakdown"
            :key="mode"
            class="pro-bar-row"
          >
            <span>{{ billingModeLabel(String(mode)) }}</span>
            <div class="pro-bar-track">
              <div
                class="pro-bar-fill"
                :style="{ width: barWidth(count, modeMax) }"
              />
            </div>
            <span>{{ count }}</span>
          </div>
        </div>
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { formatCurrency } = useFormatters()
const { planLabel, billingModeLabel } = useCodeLabels()

const metrics = ref<any>(null)

const planMax = computed(() =>
  Math.max(...Object.values(metrics.value?.planBreakdown ?? { _: 1 }) as number[], 1),
)
const modeMax = computed(() =>
  Math.max(...Object.values(metrics.value?.modeBreakdown ?? { _: 1 }) as number[], 1),
)

function barWidth(count: number, max: number) {
  return `${Math.round((count / max) * 100)}%`
}

onMounted(async () => {
  const res: any = await $fetch('/api/admin/metrics')
  metrics.value = res.data
})
</script>
