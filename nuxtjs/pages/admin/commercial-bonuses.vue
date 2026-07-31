<template>
  <div data-testid="admin-commercial-bonuses-page">
    <ProPageHeader
      :title="$t('admin.commercialBonuses.title')"
      :subtitle="$t('admin.commercialBonuses.subtitle')"
    />

    <ProCard class="pf-bonus-toolbar">
      <div class="pro-list-toolbar__filters pro-field-inline-wrap">
        <div class="pf-bonus-period-nav" data-testid="bonus-period-nav">
          <ProButton
            variant="secondary"
            :disabled="!canPrevPeriod"
            test-id="bonus-period-prev"
            @click="shiftPeriod(-1)"
          >
            <ProIcon name="chevron_left" />
          </ProButton>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="bonus-period">{{ $t('admin.commercialBonuses.filterPeriod') }}</label>
            <select
              id="bonus-period"
              v-model="periodYm"
              class="pro-input"
              data-testid="bonus-filter-period"
              @change="load"
            >
              <option v-for="p in periods" :key="p" :value="p">{{ p }}</option>
            </select>
          </div>
          <ProButton
            variant="secondary"
            :disabled="!canNextPeriod"
            test-id="bonus-period-next"
            @click="shiftPeriod(1)"
          >
            <ProIcon name="chevron_right" />
          </ProButton>
        </div>

        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="bonus-trend-months">{{ $t('admin.commercialBonuses.filterTrend') }}</label>
          <select
            id="bonus-trend-months"
            v-model.number="trendMonths"
            class="pro-input"
            data-testid="bonus-filter-trend"
            @change="load"
          >
            <option :value="3">{{ $t('admin.commercialBonuses.trendMonths', { n: 3 }) }}</option>
            <option :value="6">{{ $t('admin.commercialBonuses.trendMonths', { n: 6 }) }}</option>
            <option :value="12">{{ $t('admin.commercialBonuses.trendMonths', { n: 12 }) }}</option>
          </select>
        </div>

        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="bonus-status">{{ $t('admin.commercialBonuses.filterStatus') }}</label>
          <select id="bonus-status" v-model="statusFilter" class="pro-input" data-testid="bonus-filter-status" @change="load">
            <option value="">{{ $t('admin.commercialBonuses.filterAll') }}</option>
            <option value="in_progress">{{ $t('commissionSheet.status.in_progress') }}</option>
            <option value="earned">{{ $t('commissionSheet.status.earned') }}</option>
            <option value="paid">{{ $t('commissionSheet.status.paid') }}</option>
          </select>
        </div>

        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="bonus-commercial">{{ $t('admin.commercialBonuses.filterCommercial') }}</label>
          <select
            id="bonus-commercial"
            v-model="commercialFilter"
            class="pro-input"
            data-testid="bonus-filter-commercial"
            @change="load"
          >
            <option value="">{{ $t('admin.commercialBonuses.filterAll') }}</option>
            <option v-for="c in commercials" :key="c.userId" :value="c.userId">
              {{ c.fullName }} ({{ c.email }})
            </option>
          </select>
        </div>
      </div>
    </ProCard>

    <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

    <div class="pro-grid-kpi pro-mt-lg" data-testid="bonus-kpi-row">
      <ProKpi :label="$t('admin.commercialBonuses.kpiPeriod')" :value="periodYm" />
      <ProKpi :label="$t('admin.commercialBonuses.kpiDue')" :value="formatCurrency(kpi.dueCents)" />
      <ProKpi :label="$t('admin.commercialBonuses.kpiPaid')" :value="formatCurrency(kpi.paidCents)" />
      <ProKpi :label="$t('admin.commercialBonuses.kpiMet')" :value="String(kpi.metCount)" />
      <ProKpi :label="$t('admin.commercialBonuses.kpiInProgress')" :value="String(kpi.inProgressCount)" />
      <ProKpi
        :label="$t('admin.commercialBonuses.kpiAvgMix')"
        :value="`${kpi.avgMixPct}%`"
      />
    </div>

    <ProCard
      class="pro-mt-lg"
      :title="$t('admin.commercialBonuses.trendTitle')"
      data-testid="bonus-trend-card"
    >
      <p class="pf-bonus-hint">
        {{ $t('admin.commercialBonuses.trendHint', { target: targetPct }) }}
      </p>
      <div v-if="trend.length" class="pf-trend-chart" role="img" :aria-label="$t('admin.commercialBonuses.trendTitle')">
        <div class="pf-trend-chart__plot">
          <div
            v-for="pt in trend"
            :key="pt.periodYm"
            class="pf-trend-col"
            :class="{ 'pf-trend-col--selected': pt.periodYm === periodYm }"
            :data-testid="`bonus-trend-col-${pt.periodYm}`"
            @click="selectPeriod(pt.periodYm)"
          >
            <div class="pf-trend-col__bars">
              <div
                class="pf-trend-chart__threshold"
                :style="{ bottom: `${thresholdBottomPct}%` }"
                aria-hidden="true"
              />
              <div
                class="pf-trend-bar pf-trend-bar--mix"
                :style="{ height: `${mixBarHeight(pt.avgMixPct)}%` }"
                :title="$t('admin.commercialBonuses.trendMixTooltip', { pct: pt.avgMixPct, period: pt.periodYm })"
              />
              <div
                class="pf-trend-bar pf-trend-bar--met"
                :style="{ height: `${metBarHeight(pt.metCount)}%` }"
                :title="$t('admin.commercialBonuses.trendMetTooltip', { count: pt.metCount, period: pt.periodYm })"
              />
            </div>
            <div class="pf-trend-col__label">{{ shortPeriod(pt.periodYm) }}</div>
            <div class="pf-trend-col__meta">{{ pt.avgMixPct }}%</div>
          </div>
        </div>
        <ul class="pf-chart-legend" :aria-label="$t('admin.commercialBonuses.legend')">
          <li><span class="pf-swatch pf-swatch--mix" aria-hidden="true" />{{ $t('admin.commercialBonuses.legendAvgMix') }}</li>
          <li><span class="pf-swatch pf-swatch--met" aria-hidden="true" />{{ $t('admin.commercialBonuses.legendMet') }}</li>
          <li><span class="pf-swatch pf-swatch--threshold" aria-hidden="true" />{{ $t('admin.commercialBonuses.thresholdLabel', { target: targetPct }) }}</li>
        </ul>
      </div>
      <p v-else class="text-muted">{{ $t('admin.commercialBonuses.trendEmpty') }}</p>
    </ProCard>

    <div class="pro-grid-2 pro-mt-lg">
      <ProCard :title="$t('admin.commercialBonuses.compareTitle')" data-testid="bonus-compare-card">
        <p class="pf-bonus-hint">{{ $t('admin.commercialBonuses.compareHint', { period: periodYm, target: targetPct }) }}</p>
        <div v-if="comparison.length" class="pro-bar-chart pf-compare-chart">
          <div
            v-for="row in comparison"
            :key="row.commercialUserId"
            class="pro-bar-row pf-compare-row"
            :data-testid="`bonus-compare-${row.commercialUserId}`"
          >
            <span class="pf-compare-name" :title="row.commercialEmail">{{ row.commercialFullName }}</span>
            <div class="pro-bar-track pf-compare-track">
              <div
                class="pro-bar-fill"
                :class="`pf-compare-fill--${row.status}`"
                :style="{ width: `${Math.min(row.mixPct, 100)}%` }"
              />
              <div
                class="pf-compare-threshold"
                :style="{ left: `${targetPct}%` }"
                aria-hidden="true"
              />
            </div>
            <span class="pf-compare-pct">{{ row.mixPct }}%</span>
          </div>
        </div>
        <p v-else class="text-muted">{{ $t('admin.commercialBonuses.compareEmpty') }}</p>
      </ProCard>

      <ProCard :title="$t('admin.commercialBonuses.monthDetailTitle', { period: periodYm })">
        <p class="pf-bonus-hint">{{ $t('admin.commercialBonuses.monthDetailHint') }}</p>
        <ul class="pf-month-stats" data-testid="bonus-month-stats">
          <li>{{ $t('admin.commercialBonuses.statEarned', { count: kpi.earnedCount }) }}</li>
          <li>{{ $t('admin.commercialBonuses.statPaid', { count: kpi.paidCount }) }}</li>
          <li>{{ $t('admin.commercialBonuses.statInProgress', { count: kpi.inProgressCount }) }}</li>
          <li>{{ $t('admin.commercialBonuses.statDue', { amount: formatCurrency(kpi.dueCents) }) }}</li>
        </ul>
      </ProCard>
    </div>

    <ProCard class="pro-mt-lg" :title="$t('admin.commercialBonuses.tableTitle', { period: periodYm })">
      <ProTable :empty="!items.length" :empty-title="$t('admin.commercialBonuses.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commercialBonuses.colCommercial') }}</th>
            <th>{{ $t('admin.commercialBonuses.colProgress') }}</th>
            <th>{{ $t('admin.commercialBonuses.colActivations') }}</th>
            <th>{{ $t('admin.commercialBonuses.colAmount') }}</th>
            <th>{{ $t('admin.commercialBonuses.colStatus') }}</th>
            <th>{{ $t('admin.commercialBonuses.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(row, idx) in items"
            :key="row.awardId || `${row.commercialUserId}-${row.periodYm}-${idx}`"
          >
            <td>
              <div>{{ row.commercialFullName }}</div>
              <div class="text-muted">{{ row.commercialEmail }}</div>
            </td>
            <td>
              <div class="pf-progress-cell">
                <div class="pro-bar-track pf-progress-track">
                  <div
                    class="pro-bar-fill"
                    :class="`pf-compare-fill--${row.status}`"
                    :style="{ width: `${Math.min(row.progress, 100)}%` }"
                  />
                </div>
                <span>{{ row.progress }}% / {{ row.target }}%</span>
              </div>
            </td>
            <td>{{ row.triennialCount ?? 0 }}/{{ row.subCount ?? 0 }}</td>
            <td>{{ formatCurrency(row.amountCents) }}</td>
            <td>
              <ProBadge :variant="statusVariant(row.status)">
                {{ $t(`commissionSheet.status.${row.status}`) }}
              </ProBadge>
            </td>
            <td>
              <ProButton
                v-if="row.status === 'earned' && row.awardId"
                variant="secondary"
                :loading="payingId === row.awardId"
                :test-id="`bonus-mark-paid-${row.awardId}`"
                @click="markPaid(row)"
              >
                {{ $t('admin.commercialBonuses.markPaid') }}
              </ProButton>
              <span v-else class="text-muted">—</span>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-only' })

const { t } = useI18n()
const { formatCurrency } = useFormatters()
const { mapError } = useApiError()

type BonusKpi = {
  earnedCount: number
  paidCount: number
  inProgressCount: number
  dueCents: number
  paidCents: number
  avgMixPct: number
  metCount: number
}

type TrendPoint = {
  periodYm: string
  earnedCount: number
  paidCount: number
  inProgressCount: number
  metCount: number
  dueCents: number
  paidCents: number
  avgMixPct: number
  activeCommercials: number
}

type CompareRow = {
  commercialUserId: string
  commercialFullName: string
  commercialEmail: string
  mixPct: number
  triennialCount: number
  subCount: number
  status: string
  amountCents: number
  awardId?: string
}

const items = ref<any[]>([])
const commercials = ref<any[]>([])
const periods = ref<string[]>([])
const trend = ref<TrendPoint[]>([])
const comparison = ref<CompareRow[]>([])
const kpi = ref<BonusKpi>({
  earnedCount: 0,
  paidCount: 0,
  inProgressCount: 0,
  dueCents: 0,
  paidCents: 0,
  avgMixPct: 0,
  metCount: 0,
})
const periodYm = ref('')
const statusFilter = ref('')
const commercialFilter = ref('')
const trendMonths = ref(6)
const targetPct = ref(55)
const error = ref('')
const payingId = ref('')

const canPrevPeriod = computed(() => {
  const idx = periods.value.indexOf(periodYm.value)
  return idx >= 0 && idx < periods.value.length - 1
})

const canNextPeriod = computed(() => {
  const idx = periods.value.indexOf(periodYm.value)
  return idx > 0
})

const maxMetCount = computed(() =>
  Math.max(1, ...trend.value.map((p) => p.metCount), 0),
)

const thresholdBottomPct = computed(() => Math.min(100, Math.max(0, targetPct.value)))

function statusVariant(status: string): 'success' | 'warning' | 'neutral' {
  if (status === 'paid' || status === 'earned') return 'success'
  if (status === 'in_progress') return 'warning'
  return 'neutral'
}

function shortPeriod(ym: string) {
  if (!ym || ym.length < 7) return ym
  return ym.slice(5)
}

function mixBarHeight(pct: number) {
  return Math.min(100, Math.max(4, pct))
}

function metBarHeight(count: number) {
  return Math.min(100, Math.max(count > 0 ? 8 : 0, (count / maxMetCount.value) * 100))
}

function selectPeriod(ym: string) {
  if (!ym || ym === periodYm.value) return
  periodYm.value = ym
  void load()
}

function shiftPeriod(delta: number) {
  const idx = periods.value.indexOf(periodYm.value)
  if (idx < 0) return
  // periods are DESC: index+1 = older, index-1 = newer
  const next = periods.value[idx - delta]
  if (!next) return
  periodYm.value = next
  void load()
}

async function load() {
  error.value = ''
  try {
    const query: Record<string, string> = {
      trendMonths: String(trendMonths.value),
    }
    if (periodYm.value) query.periodYm = periodYm.value
    if (statusFilter.value) query.status = statusFilter.value
    if (commercialFilter.value) query.commercialId = commercialFilter.value
    const res: any = await $fetch('/api/admin/commercial-bonuses', { query })
    const data = res.data ?? res
    periodYm.value = data.periodYm || periodYm.value
    periods.value = data.periods ?? []
    trend.value = data.trend ?? []
    comparison.value = data.comparison ?? []
    items.value = data.items ?? []
    targetPct.value = data.targetPct ?? 55
    kpi.value = {
      earnedCount: data.kpi?.earnedCount ?? 0,
      paidCount: data.kpi?.paidCount ?? 0,
      inProgressCount: data.kpi?.inProgressCount ?? 0,
      dueCents: data.kpi?.dueCents ?? 0,
      paidCents: data.kpi?.paidCents ?? 0,
      avgMixPct: data.kpi?.avgMixPct ?? 0,
      metCount: data.kpi?.metCount ?? 0,
    }
  } catch (e: any) {
    error.value = mapError(e)
  }
}

async function markPaid(row: any) {
  if (!row.awardId) return
  if (!confirm(t('admin.commercialBonuses.confirmMarkPaid'))) return
  payingId.value = row.awardId
  error.value = ''
  try {
    await $fetch(`/api/admin/commercial-bonuses/${row.awardId}/mark-paid`, { method: 'POST' })
    await load()
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    payingId.value = ''
  }
}

onMounted(async () => {
  const commercialsRes: any = await $fetch('/api/admin/commercials').catch(() => null)
  commercials.value = commercialsRes?.data ?? commercialsRes ?? []
  await load()
})
</script>

<style scoped>
.pro-field-inline-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: flex-end;
}

.pf-bonus-period-nav {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
}

.pf-bonus-hint {
  color: var(--pf-vet-muted, #6b7280);
  font-size: 0.875rem;
  margin: 0 0 1rem;
}

.text-muted {
  color: var(--pf-vet-muted, #6b7280);
  font-size: 0.875rem;
}

.pf-trend-chart__plot {
  position: relative;
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  min-height: 180px;
  padding: 0.5rem 0.25rem 0;
  border-bottom: 1px solid var(--pf-vet-border, #e5e7eb);
}

.pf-trend-col {
  flex: 1 1 0;
  min-width: 2.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  cursor: pointer;
  border-radius: 0.5rem;
  padding: 0.25rem;
  transition: background 0.15s ease;
}

.pf-trend-col:hover,
.pf-trend-col--selected {
  background: rgba(42, 157, 143, 0.08);
}

.pf-trend-col__bars {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 3px;
  width: 100%;
  height: 140px;
}

.pf-trend-chart__threshold {
  position: absolute;
  left: 0;
  right: 0;
  height: 0;
  border-top: 2px dashed var(--pf-vet-alert, #e07a5f);
  opacity: 0.7;
  pointer-events: none;
  z-index: 1;
}

.pf-trend-bar {
  width: 42%;
  max-width: 1.1rem;
  border-radius: 0.35rem 0.35rem 0 0;
  min-height: 2px;
  transition: height 0.25s ease;
}

.pf-trend-bar--mix {
  background: var(--pf-vet-accent, #2a9d8f);
}

.pf-trend-bar--met {
  background: var(--pf-vet-primary, #1d3557);
}

.pf-trend-col__label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--pf-vet-primary, #1d3557);
}

.pf-trend-col__meta {
  font-size: 0.7rem;
  color: var(--pf-vet-muted, #6b7280);
}

.pf-chart-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  list-style: none;
  margin: 1rem 0 0;
  padding: 0;
  font-size: 0.8125rem;
  color: var(--pf-vet-muted, #6b7280);
}

.pf-chart-legend li {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.pf-swatch {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 2px;
  display: inline-block;
}

.pf-swatch--mix { background: var(--pf-vet-accent, #2a9d8f); }
.pf-swatch--met { background: var(--pf-vet-primary, #1d3557); }
.pf-swatch--threshold {
  background: transparent;
  border-top: 2px dashed var(--pf-vet-alert, #e07a5f);
  height: 0;
  width: 1rem;
}

.pf-compare-chart .pf-compare-row {
  grid-template-columns: minmax(7rem, 9rem) 1fr 2.75rem;
}

.pf-compare-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pf-compare-track {
  position: relative;
}

.pf-compare-threshold {
  position: absolute;
  top: -2px;
  bottom: -2px;
  width: 0;
  border-left: 2px dashed var(--pf-vet-alert, #e07a5f);
  opacity: 0.9;
}

.pf-compare-fill--paid,
.pf-compare-fill--earned {
  background: var(--pf-vet-accent, #2a9d8f);
}

.pf-compare-fill--in_progress {
  background: var(--pf-vet-primary, #1d3557);
  opacity: 0.65;
}

.pf-compare-pct {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.pf-month-stats {
  margin: 0;
  padding-left: 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  color: var(--pf-vet-primary, #1d3557);
}

.pf-progress-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 8rem;
}

.pf-progress-track {
  width: 100%;
}
</style>
