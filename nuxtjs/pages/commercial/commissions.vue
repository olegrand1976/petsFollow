<template>
  <div data-testid="commercial-commissions-page">
    <ProPageHeader
      :title="$t('commercial.commissions.paymentsTitle')"
      :subtitle="$t('commercial.commissions.paymentsSubtitle')"
    />
    <div v-if="!hasIban" class="pro-banner">
      <span>{{ $t('commercial.commissions.ibanBanner') }}</span>
      <NuxtLink to="/commercial/settings">{{ $t('commercial.commissions.ibanBannerLink') }}</NuxtLink>
    </div>

    <div v-if="summary" class="pro-grid-kpi">
      <ProKpi :value="formatCurrency(summary.monthEarnedCents)" :label="$t('commercial.commissions.monthDue')" />
      <ProKpi :value="formatCurrency(summary.lifetimeEarnedCents)" :label="$t('commercial.commissions.lifetime')" />
      <ProKpi :value="formatCurrency(pendingPayoutCents)" :label="$t('commercial.commissions.pendingPayouts')" />
      <ProKpi :value="formatCurrency(paidPayoutCents)" :label="$t('commercial.commissions.paidPayouts')" />
    </div>

    <ProCard :title="$t('commercial.commissions.payoutsTitle')" class="pro-mt-lg">
      <ProTable :empty="!(summary?.payoutHistory || []).length" :empty-title="$t('commercial.commissions.payoutsEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.commissions.colPeriod') }}</th>
            <th>{{ $t('commercial.commissions.colAmount') }}</th>
            <th>{{ $t('commercial.commissions.colStatus') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in summary?.payoutHistory || []" :key="p.periodYm">
            <td>{{ p.periodYm }}</td>
            <td>{{ formatCurrency(p.amountCents) }}</td>
            <td>
              <ProBadge :variant="p.runStatus === 'paid' ? 'success' : 'warning'">
                {{ payoutStatusLabel(p.runStatus) }}
              </ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('commercial.commissions.gainsTitle')" class="pro-mt-lg" data-testid="commission-rates">
      <div class="pf-plan-compare">
        <div
          v-for="row in summary?.planRates || []"
          :key="row.code"
          class="pf-plan-compare__item"
          :class="{ 'pf-plan-compare__item--rec': row.recommended }"
        >
          <strong>{{ $t(`commissionSheet.plans.${row.code}`) }}</strong>
          <span>{{ formatPct(row.commercialRateBps) }} · {{ formatCurrency(row.commercialCents) }}</span>
          <ProBadge v-if="row.recommended" variant="success">{{ $t('commissionSheet.recommended') }}</ProBadge>
        </div>
      </div>
      <div class="pf-bonus-row" data-testid="commercial-bonus-cards">
        <ProCard v-for="b in commercialBonuses" :key="b.code" class="pf-bonus-card" :data-testid="`bonus-card-${b.code}`">
          <strong>{{ $t(`commissionSheet.bonusTitles.${b.code}`) }}</strong>
          <p>{{ formatCurrency(b.amountCents) }}</p>
          <p class="text-muted">{{ $t(`commissionSheet.bonusHints.${b.code}`) }}</p>
          <p v-if="b.code === 'commercial_mix' && b.periodYm" class="text-muted">
            {{ $t('commercial.commissions.bonusPeriod', { period: b.periodYm }) }}
          </p>
          <ProBadge :variant="bonusBadgeVariant(b.status)">
            {{ $t(`commissionSheet.status.${b.status || 'available'}`) }}
            <template v-if="b.progress != null && b.target">
              — {{ b.progress }}/{{ b.target }}<template v-if="b.code === 'commercial_mix'"> %</template>
            </template>
          </ProBadge>
        </ProCard>
      </div>
    </ProCard>

    <ProCard :title="$t('commercial.commissions.sheetTitle')" class="pro-mt-lg">
      <ProCommissionSheet
        audience="commercial"
        :plan-rates="summary?.planRates || []"
        :addon-rates="summary?.addonRates || []"
        :bonuses="summary?.bonuses || []"
      />
      <NuxtLink to="/commercial/pitch" class="pro-hint-link">{{ $t('commercial.commissions.pitchLink') }}</NuxtLink>
    </ProCard>

    <ProCard class="pro-mt-lg" :title="$t('commercial.commissions.ledgerTitle')">
      <div class="pf-ledger-filters pro-mb-md">
        <select v-model="ledgerType" class="pro-select" data-testid="ledger-type-filter">
          <option value="">{{ $t('commercial.commissions.typeAll') }}</option>
          <option value="subscription_pct">subscription_pct</option>
          <option value="subscription_mirror">subscription_mirror</option>
          <option value="addon_pct">addon_pct</option>
        </select>
        <ProButton
          variant="secondary"
          test-id="ledger-export-csv"
          :disabled="exporting"
          @click="exportLedgerCsv"
        >
          {{ $t('commercial.commissions.exportCsv') }}
        </ProButton>
      </div>
      <p v-if="ledgerTruncated" class="pro-hint pro-mb-md" data-testid="ledger-truncated-hint">
        {{ $t('commercial.commissions.exportCsvTruncated', { n: summary?.ledgerLimit || 50 }) }}
      </p>
      <ProTable :empty="!filteredLedger.length" :empty-title="$t('commercial.commissions.empty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.commissions.date') }}</th>
            <th>{{ $t('commercial.commissions.type') }}</th>
            <th>{{ $t('commercial.commissions.vet') }}</th>
            <th>{{ $t('commercial.commissions.client') }}</th>
            <th>{{ $t('commercial.commissions.base') }}</th>
            <th>{{ $t('commercial.commissions.amount') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in filteredLedger" :key="row.id">
            <td>{{ row.accruedAt?.substring?.(0, 10) || row.accruedAt }}</td>
            <td><ProBadge variant="neutral">{{ row.sourceType }}</ProBadge></td>
            <td>{{ row.vetEmail }}</td>
            <td>{{ row.clientEmail }}</td>
            <td>{{ formatCurrency(row.baseAmountCents) }}</td>
            <td>{{ formatCurrency(row.commissionCents) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { t } = useI18n()
const { formatCurrency } = useFormatters()
const summary = ref<any>(null)
const hasIban = ref(false)
const ledgerType = ref('')
const exporting = ref(false)

const filteredLedger = computed(() => {
  const rows = summary.value?.recentLedger || []
  if (!ledgerType.value) return rows
  return rows.filter((r: any) => r.sourceType === ledgerType.value)
})

const ledgerTruncated = computed(() => Boolean(summary.value?.ledgerTruncated))

function payoutStatusLabel(status: string) {
  const key = `commercial.commissions.runStatus.${status}`
  const label = t(key)
  return label === key ? status : label
}

const commercialBonuses = computed(() =>
  (summary.value?.bonuses || []).filter((b: any) => b.audience === 'commercial'),
)

const pendingPayoutCents = computed(() =>
  (summary.value?.payoutHistory || [])
    .filter((p: any) => p.runStatus !== 'paid')
    .reduce((sum: number, p: any) => sum + (p.amountCents || 0), 0),
)

const paidPayoutCents = computed(() =>
  (summary.value?.payoutHistory || [])
    .filter((p: any) => p.runStatus === 'paid')
    .reduce((sum: number, p: any) => sum + (p.amountCents || 0), 0),
)

function formatPct(bps: number) {
  return `${((bps || 0) / 100).toFixed(0)} %`
}

function csvEscape(v: unknown) {
  const s = String(v ?? '')
  if (/[",\n\r]/.test(s)) return `"${s.replace(/"/g, '""')}"`
  return s
}

function rowsToCsv(rows: any[]) {
  const headers = ['date', 'type', 'vet', 'client', 'baseCents', 'commissionCents', 'periodYm']
  const lines = [headers.join(',')]
  for (const row of rows) {
    lines.push([
      csvEscape(row.accruedAt?.substring?.(0, 10) || row.accruedAt),
      csvEscape(row.sourceType),
      csvEscape(row.vetEmail),
      csvEscape(row.clientEmail),
      csvEscape(row.baseAmountCents),
      csvEscape(row.commissionCents),
      csvEscape(row.periodYm),
    ].join(','))
  }
  return '\uFEFF' + lines.join('\n')
}

async function exportLedgerCsv() {
  exporting.value = true
  try {
    const res: any = await $fetch('/api/commercial/commissions', { query: { limit: 500 } })
    const data = res.data ?? res
    let rows = data?.recentLedger || []
    if (ledgerType.value) {
      rows = rows.filter((r: any) => r.sourceType === ledgerType.value)
    }
    if (!rows.length) return
    const blob = new Blob([rowsToCsv(rows)], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `commissions-ledger-${data?.monthPeriodYm || summary.value?.monthPeriodYm || 'export'}.csv`
    a.click()
    URL.revokeObjectURL(url)
    if (data?.ledgerTruncated) {
      // keep UI hint in sync if export hit the cap
      summary.value = { ...summary.value, ledgerTruncated: true, ledgerLimit: data.ledgerLimit || 500 }
    }
  } finally {
    exporting.value = false
  }
}

function bonusBadgeVariant(status?: string): 'success' | 'warning' | 'neutral' {
  if (status === 'paid' || status === 'earned') return 'success'
  if (status === 'in_progress') return 'warning'
  return 'neutral'
}

onMounted(async () => {
  const [commRes, profileRes]: any[] = await Promise.all([
    $fetch('/api/commercial/commissions'),
    $fetch('/api/commercial/me/payout-profile').catch(() => null),
  ])
  summary.value = commRes.data ?? commRes
  const profile = profileRes?.data ?? profileRes
  hasIban.value = Boolean(profile?.iban || profile?.payoutIban)
})
</script>

<style scoped>
.pro-grid-kpi {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.75rem;
}
.pro-mt-lg { margin-top: 1.25rem; }
.pro-mt-md { margin-top: 1rem; }
.pro-banner {
  margin-bottom: 1rem;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
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
.pf-plan-compare__item--rec { border-color: var(--pf-vet-accent); }
.pf-bonus-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}
.pf-bonus-card p { margin: 0.25rem 0; }
.pro-hint-link {
  display: inline-block;
  margin-top: 0.75rem;
  color: var(--pf-vet-accent);
}
.pf-ledger-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-end;
  max-width: none;
}
@media (max-width: 900px) {
  .pf-plan-compare { grid-template-columns: 1fr; }
}
</style>
