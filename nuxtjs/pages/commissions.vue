<template>
  <div data-testid="vet-commissions-page">
    <ProPageHeader
      :title="$t('commissions.title')"
      :subtitle="$t('commissions.subtitle')"
    />

    <p class="pro-hint pro-mb" data-testid="commissions-payout-schedule">
      {{ $t('commissions.payoutSchedule') }}
    </p>

    <div
      v-if="!summary.payoutProfileComplete"
      class="pf-commissions-banner pro-mb"
      data-testid="commissions-payout-incomplete"
    >
      <ProIcon name="account_balance" />
      <div>
        <p>{{ $t('commissions.payoutIncompleteBanner') }}</p>
        <NuxtLink to="/settings" class="pf-commissions-banner__link">
          {{ $t('commissions.payoutIncompleteLink') }}
        </NuxtLink>
      </div>
    </div>

    <div class="pro-kpi-grid">
      <ProKpi :label="$t('commissions.kpiMonth')" :value="formatCurrency(summary.monthEarnedCents ?? 0)" />
      <ProKpi :label="$t('commissions.kpiLifetime')" :value="formatCurrency(summary.lifetimeEarnedCents ?? 0)" />
      <ProKpi :label="$t('commissions.kpiClients')" :value="String(summary.eligibleClients ?? 0)" />
      <ProKpi :label="$t('commissions.kpiRateHeart')" :value="heartRateLabel" />
    </div>

    <p v-if="summary.nextTierMinClients" class="pro-hint pro-mb">
      {{ $t('commissions.nextTier', { n: summary.nextTierMinClients }) }}
    </p>

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

    <ProCard :title="$t('commissions.payoutsTitle')" class="pro-mb">
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
              <ProBadge :variant="lineStatusVariant(p.status)">
                {{ lineStatusLabel(p.status) }}
              </ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <details class="pf-commissions-details">
      <summary>{{ $t('commissions.detailsSummary') }}</summary>
      <div class="pf-commissions-details__body">
        <ProCommissionSheet
          audience="vet"
          :plan-rates="summary.planRates || []"
          :addon-rates="summary.addonRates || []"
          :bonuses="vetBonuses"
        />
      </div>
    </details>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const { t } = useI18n()
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
  payoutProfileComplete: true,
})

const heartRateLabel = computed(() =>
  `${((summary.value.heartRateBps ?? summary.value.currentRateBps ?? 0) / 100).toFixed(0)}%`,
)

const vetBonuses = computed(() =>
  (summary.value.bonuses || []).filter((b: any) => !b.audience || b.audience === 'vet'),
)

function lineStatusLabel(status: string) {
  const key = `commissions.lineStatus.${status}`
  const label = t(key)
  return label === key ? status : label
}

function lineStatusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'paid':
      return 'success'
    case 'ready_to_pay':
      return 'warning'
    case 'missing_info':
      return 'danger'
    case 'accruing':
      return 'neutral'
    default:
      return 'neutral'
  }
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
.pro-mb { margin-bottom: 1.25rem; }
.pro-hint {
  margin: 0 0 1.25rem;
  color: var(--pf-vet-accent);
  font-size: 0.9rem;
}
.pf-commissions-banner {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  padding: 0.85rem 1rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
}
.pf-commissions-banner p {
  margin: 0 0 0.35rem;
}
.pf-commissions-banner__link {
  color: var(--pf-vet-accent);
  font-weight: 600;
  text-decoration: none;
}
.pf-commissions-banner__link:hover {
  text-decoration: underline;
}
.pf-commissions-details {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
  padding: 0.75rem 1rem;
}
.pf-commissions-details > summary {
  cursor: pointer;
  font-weight: 600;
  color: var(--pf-vet-primary);
  list-style: none;
}
.pf-commissions-details > summary::-webkit-details-marker {
  display: none;
}
.pf-commissions-details > summary::before {
  content: '▸';
  display: inline-block;
  margin-right: 0.4rem;
  transition: transform 0.15s ease;
}
.pf-commissions-details[open] > summary::before {
  transform: rotate(90deg);
}
.pf-commissions-details__body {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--pf-vet-border);
}
@media (max-width: 900px) {
  .pro-kpi-grid { grid-template-columns: 1fr 1fr; }
}
</style>
