<template>
  <div data-testid="vet-commissions-page">
    <ProPageHeader
      :title="$t('commissions.title')"
      :subtitle="$t('commissions.subtitle')"
    />

    <div class="pro-kpi-grid">
      <ProKpi :label="$t('commissions.kpiClients')" :value="String(summary.eligibleClients ?? 0)" />
      <ProKpi :label="$t('commissions.kpiRate')" :value="rateLabel" />
      <ProKpi :label="$t('commissions.kpiMonth')" :value="formatCurrency(summary.monthEarnedCents ?? 0)" />
      <ProKpi :label="$t('commissions.kpiLifetime')" :value="formatCurrency(summary.lifetimeEarnedCents ?? 0)" />
    </div>

    <ProCard :title="$t('commissions.tiersTitle')" class="pro-mb">
      <ul class="pro-commissions-tiers">
        <li v-for="t in summary.tiers || []" :key="t.minClients">
          {{ formatTier(t) }}
        </li>
      </ul>
      <p v-if="summary.nextTierMinClients" class="pro-hint">
        {{ $t('commissions.nextTier', { n: summary.nextTierMinClients }) }}
      </p>
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

const { t } = useI18n()
const { formatCurrency } = useFormatters()

const summary = ref<any>({
  eligibleClients: 0,
  currentRateBps: 500,
  monthEarnedCents: 0,
  lifetimeEarnedCents: 0,
  tiers: [],
  recentLedger: [],
  payoutHistory: [],
})

const rateLabel = computed(() => `${((summary.value.currentRateBps || 0) / 100).toFixed(0)}%`)

function formatTier(tier: any) {
  const max = tier.maxClients == null ? '+' : `–${tier.maxClients}`
  const pct = (tier.rateBps / 100).toFixed(0)
  return t('commissions.tierLine', { min: tier.minClients, max, pct })
}

onMounted(async () => {
  const res: any = await $fetch('/api/vet/commissions')
  summary.value = res.data ?? res
})
</script>

<style scoped>
.pro-kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.pro-mb {
  margin-bottom: 1.25rem;
}

.pro-commissions-tiers {
  margin: 0;
  padding-left: 1.2rem;
  color: var(--pf-vet-text-muted);
}

.pro-hint {
  margin: 0.75rem 0 0;
  color: var(--pf-vet-accent);
  font-size: 0.9rem;
}

@media (max-width: 900px) {
  .pro-kpi-grid {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
