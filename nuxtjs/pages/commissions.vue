<template>
  <div data-testid="vet-commissions-page">
    <ProPageHeader
      :title="$t('commissions.title')"
      :subtitle="$t('commissions.subtitle')"
    />

    <div class="pro-kpi-grid">
      <ProKpi :label="$t('commissions.kpiClients')" :value="String(summary.eligibleClients ?? 0)" />
      <ProKpi :label="$t('commissions.kpiRateBase')" :value="baseRateLabel" />
      <ProKpi :label="$t('commissions.kpiRateHeart')" :value="heartRateLabel" />
      <ProKpi :label="$t('commissions.kpiMonth')" :value="formatCurrency(summary.monthEarnedCents ?? 0)" />
      <ProKpi :label="$t('commissions.kpiLifetime')" :value="formatCurrency(summary.lifetimeEarnedCents ?? 0)" />
    </div>

    <ProCard :title="$t('commissions.gainsTitle')" class="pro-mb">
      <div class="pf-plan-compare">
        <div
          v-for="row in summary.planRates || []"
          :key="row.code"
          class="pf-plan-compare__item"
          :class="{ 'pf-plan-compare__item--rec': row.recommended }"
        >
          <strong>{{ $t(`commissionSheet.plans.${row.code}`) }}</strong>
          <span>{{ formatPct(row.vetRateBpsMax) }} · {{ formatCurrency(row.vetCentsMax) }}</span>
          <ProBadge v-if="row.recommended" variant="success">{{ $t('commissionSheet.recommended') }}</ProBadge>
        </div>
      </div>
      <div v-if="(summary.bonuses || []).length" class="pf-bonus-row">
        <ProCard v-for="b in summary.bonuses" :key="b.code" class="pf-bonus-card">
          <strong>{{ $t(`commissionSheet.bonusTitles.${b.code}`) }}</strong>
          <p>{{ formatCurrency(b.amountCents) }}</p>
          <p class="text-muted">{{ $t(`commissionSheet.bonusHints.${b.code}`) }}</p>
          <ProBadge :variant="b.status === 'earned' ? 'success' : b.status === 'in_progress' ? 'warning' : 'neutral'">
            {{ $t(`commissionSheet.status.${b.status || 'available'}`) }}
            <template v-if="b.progress != null && b.target"> — {{ b.progress }}/{{ b.target }}</template>
          </ProBadge>
        </ProCard>
      </div>
      <p v-if="summary.nextTierMinClients" class="pro-hint">
        {{ $t('commissions.nextTier', { n: summary.nextTierMinClients }) }}
      </p>
    </ProCard>

    <ProCard :title="$t('commissions.sheetTitle')" class="pro-mb">
      <ProCommissionSheet
        audience="vet"
        :plan-rates="summary.planRates || []"
        :bonuses="summary.bonuses || []"
      />
    </ProCard>

    <ProCard :title="$t('commissions.ledgerTitle')" class="pro-mb">
      <ProTable :empty="!(summary.recentLedger || []).length" :empty-title="$t('commissions.ledgerEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commissions.colDate') }}</th>
            <th>{{ $t('commissions.colClient') }}</th>
            <th>{{ $t('commissions.colPet') }}</th>
            <th>{{ $t('commissions.colBase') }}</th>
            <th>{{ $t('commissions.colRate') }}</th>
            <th>{{ $t('commissions.colCommission') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in summary.recentLedger || []" :key="row.id">
            <td>{{ row.accruedAt?.substring?.(0, 10) || row.periodYm }}</td>
            <td>{{ row.clientEmail }}</td>
            <td>{{ row.petName }}</td>
            <td>{{ formatCurrency(row.baseAmountCents) }}</td>
            <td>{{ (row.rateBps / 100).toFixed(0) }}%</td>
            <td>{{ formatCurrency(row.commissionCents) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('commissions.payoutsTitle')">
      <ProTable :empty="!(summary.payoutHistory || []).length" :empty-title="$t('commissions.payoutsEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commissions.colPeriod') }}</th>
            <th>{{ $t('commissions.colAmount') }}</th>
            <th>{{ $t('commissions.colStatus') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in summary.payoutHistory || []" :key="p.periodYm">
            <td>{{ p.periodYm }}</td>
            <td>{{ formatCurrency(p.amountCents) }}</td>
            <td>
              <ProBadge :variant="p.runStatus === 'paid' ? 'success' : 'warning'">
                {{ p.runStatus }}
              </ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const { formatCurrency } = useFormatters()

const summary = ref<any>({
  eligibleClients: 0,
  currentRateBps: 700,
  monthEarnedCents: 0,
  lifetimeEarnedCents: 0,
  tiers: [],
  planRates: [],
  bonuses: [],
  recentLedger: [],
  payoutHistory: [],
})

const baseRateLabel = computed(() =>
  `${((summary.value.currentBaseRateBps ?? summary.value.currentRateBps ?? 0) / 100).toFixed(0)}%`,
)
const heartRateLabel = computed(() =>
  `${((summary.value.heartRateBps ?? summary.value.currentRateBps ?? 0) / 100).toFixed(0)}%`,
)

function formatPct(bps: number) {
  return `${((bps || 0) / 100).toFixed(0)} %`
}

onMounted(async () => {
  const res: any = await $fetch('/api/vet/commissions')
  summary.value = res.data ?? res
})
</script>

<style scoped>
.pro-kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 1rem;
  margin-bottom: 1.25rem;
}
.pro-mb { margin-bottom: 1.25rem; }
.pro-hint {
  margin: 0.75rem 0 0;
  color: var(--pf-vet-accent);
  font-size: 0.9rem;
}
.pf-plan-compare {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
}
.pf-plan-compare__item {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  padding: 0.75rem;
  display: grid;
  gap: 0.35rem;
}
.pf-plan-compare__item--rec {
  border-color: var(--pf-vet-accent);
}
.pf-bonus-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}
.pf-bonus-card p { margin: 0.25rem 0; }
@media (max-width: 900px) {
  .pro-kpi-grid, .pf-plan-compare { grid-template-columns: 1fr 1fr; }
}
</style>
