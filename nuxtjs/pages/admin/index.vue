<template>
  <div>
    <ProPageHeader
      title="Tableau de bord admin"
      subtitle="KPIs commercial, marketing et facturation."
    />
    <div v-if="metrics" class="pro-grid-kpi">
      <ProKpi :value="formatEur(metrics.totalRevenueCents)" label="CA total" />
      <ProKpi :value="formatEur(metrics.periodRevenueCents)" label="CA 30 jours" />
      <ProKpi :value="formatEur(metrics.mrrCents)" label="MRR abonnements" />
      <ProKpi :value="metrics.userCount" label="Inscriptions" />
      <ProKpi :value="metrics.petCount" label="Animaux" />
      <ProKpi
        :value="`${metrics.conversionRatePercent.toFixed(1)}%`"
        label="Conversion paiement"
      />
      <ProKpi :value="metrics.pendingPayments" label="Paiements en attente" />
      <ProKpi :value="metrics.pastDueCount" label="Impayés" />
    </div>
    <div v-if="metrics" class="pro-grid-2 pro-mt-lg">
      <ProCard title="Répartition plans">
        <div class="pro-bar-chart">
          <div
            v-for="(count, plan) in metrics.planBreakdown"
            :key="plan"
            class="pro-bar-row"
          >
            <span>{{ plan }}</span>
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
      <ProCard title="Modes de facturation">
        <div class="pro-bar-chart">
          <div
            v-for="(count, mode) in metrics.modeBreakdown"
            :key="mode"
            class="pro-bar-row"
          >
            <span>{{ mode }}</span>
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

const metrics = ref<any>(null)

const planMax = computed(() =>
  Math.max(...Object.values(metrics.value?.planBreakdown ?? { _: 1 }) as number[], 1),
)
const modeMax = computed(() =>
  Math.max(...Object.values(metrics.value?.modeBreakdown ?? { _: 1 }) as number[], 1),
)

function formatEur(cents: number) {
  return `${(cents / 100).toFixed(2)} €`
}

function barWidth(count: number, max: number) {
  return `${Math.round((count / max) * 100)}%`
}

onMounted(async () => {
  const res: any = await $fetch('/api/admin/metrics')
  metrics.value = res.data
})
</script>
