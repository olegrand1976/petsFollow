<template>
  <div data-testid="admin-commercial-commissions-page">
    <ProPageHeader
      :title="$t('admin.commercialCommissions.title')"
      :subtitle="$t('admin.commercialCommissions.subtitle')"
    />

    <div class="pro-kpi-grid">
      <ProKpi :label="$t('admin.commercialCommissions.kpiPeriod')" :value="periodYm" />
      <ProKpi :label="$t('admin.commercialCommissions.kpiTotal')" :value="formatCurrency(totalCents)" />
      <ProKpi :label="$t('admin.commercialCommissions.kpiStatus')" :value="runStatusLabel" />
      <ProKpi :label="$t('admin.commercialCommissions.kpiRate')" :value="`${(commercialRateBps / 100).toFixed(0)}%`" />
    </div>

    <ProCard :title="$t('admin.commercialCommissions.periodTitle')">
      <div class="pro-list-toolbar__filters pro-field-inline-wrap">
        <div class="pro-tabs">
          <ProButton
            v-for="tab in tabs"
            :key="tab"
            :variant="viewTab === tab ? 'primary' : 'secondary'"
            :disabled="!tabHasRuns(tab)"
            @click="selectTab(tab)"
          >
            {{ $t(`admin.commercialCommissions.tab.${tab}`) }}
            <span v-if="tabCount(tab)" class="pro-tab-count">({{ tabCount(tab) }})</span>
          </ProButton>
        </div>
        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="com-period-ym">{{ $t('admin.commercialCommissions.period') }}</label>
          <select id="com-period-ym" v-model="periodYm" class="pro-input" @change="loadPeriod">
            <option v-for="p in periodsForTab" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <ProButton
          variant="secondary"
          :disabled="runStatus !== 'open' || closing"
          :loading="closing"
          test-id="commercial-commissions-close"
          @click="closePeriod"
        >
          {{ $t('admin.commercialCommissions.close') }}
        </ProButton>
        <ProButton
          :disabled="runStatus !== 'closed' || marking"
          :loading="marking"
          test-id="commercial-commissions-mark-paid"
          @click="markPaid"
        >
          {{ $t('admin.commercialCommissions.markPaid') }}
        </ProButton>
      </div>

      <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

      <p class="pro-recap">
        {{ $t('admin.commercialCommissions.recapGlobal', { amount: formatCurrency(totalCents), count: lines.length }) }}
      </p>

      <ProTable :empty="!lines.length" :empty-title="$t('admin.commercialCommissions.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commercialCommissions.colCommercial') }}</th>
            <th>{{ $t('admin.commercialCommissions.colLedger') }}</th>
            <th>{{ $t('admin.commercialCommissions.colAmount') }}</th>
            <th>{{ $t('admin.commercialCommissions.colIban') }}</th>
            <th>{{ $t('admin.commercialCommissions.colHolder') }}</th>
            <th>{{ $t('admin.commercialCommissions.colStatus') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="l in lines" :key="l.commercialUserId + (l.id || '')">
            <td>
              <div>{{ l.commercialFullName }}</div>
              <div class="text-muted">{{ l.commercialEmail }}</div>
            </td>
            <td>{{ l.ledgerCount }}</td>
            <td>{{ formatCurrency(l.amountCents) }}</td>
            <td class="pro-mono">{{ l.payoutIban || '—' }}</td>
            <td>{{ l.payoutAccountHolder || '—' }}</td>
            <td>
              <ProBadge :variant="lineVariant(effectiveStatus(l))">
                {{ statusLabel(effectiveStatus(l)) }}
              </ProBadge>
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

type Tab = 'forecast' | 'toPay' | 'paid'
const tabs: Tab[] = ['forecast', 'toPay', 'paid']

const now = new Date()
const currentPeriodYm = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
const periodYm = ref(currentPeriodYm)
const lines = ref<any[]>([])
const totalCents = ref(0)
const runStatus = ref('open')
const commercialRateBps = ref(1200)
const runs = ref<any[]>([])
const closing = ref(false)
const marking = ref(false)
const error = ref('')
const viewTab = ref<Tab>('forecast')

function tabStatus(tab: Tab): string {
  switch (tab) {
    case 'forecast':
      return 'open'
    case 'toPay':
      return 'closed'
    case 'paid':
      return 'paid'
    default: {
      const _exhaustive: never = tab
      return _exhaustive
    }
  }
}

function tabHasRuns(tab: Tab): boolean {
  if (tab === 'forecast') return true
  return runs.value.some(r => r.status === tabStatus(tab))
}

function tabCount(tab: Tab): number {
  if (tab === 'forecast') {
    return runs.value.filter(r => r.status === 'open').length || 1
  }
  return runs.value.filter(r => r.status === tabStatus(tab)).length
}

const periodsForTab = computed(() => {
  const status = tabStatus(viewTab.value)
  const fromRuns = runs.value
    .filter(r => r.status === status)
    .map((r: any) => r.periodYm as string)
  if (viewTab.value === 'forecast' && !fromRuns.includes(currentPeriodYm)) {
    return [currentPeriodYm, ...fromRuns]
  }
  return fromRuns.length ? fromRuns : [periodYm.value]
})

const runStatusLabel = computed(() => statusLabel(runStatus.value))

function statusLabel(status: string) {
  switch (status) {
    case 'open':
      return t('admin.commercialCommissions.statusOpen')
    case 'closed':
      return t('admin.commercialCommissions.statusClosed')
    case 'paid':
      return t('admin.commercialCommissions.statusPaid')
    case 'pending':
      return t('admin.commercialCommissions.statusClosed')
    default:
      return status
  }
}

function effectiveStatus(line: any) {
  return line.status === 'pending' && runStatus.value !== 'open'
    ? runStatus.value
    : (line.status || runStatus.value)
}

function lineVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  if (status === 'paid') return 'success'
  if (status === 'closed' || status === 'pending') return 'warning'
  return 'neutral'
}

async function selectTab(tab: Tab) {
  viewTab.value = tab
  const status = tabStatus(tab)
  if (tab === 'forecast') {
    periodYm.value = currentPeriodYm
  } else {
    const match = runs.value.find(r => r.status === status)
    if (match) periodYm.value = match.periodYm
  }
  await loadPeriod({ syncTab: false })
}

async function loadRunsMeta() {
  const res: any = await $fetch('/api/admin/commercial-commissions/runs')
  const data = res.data ?? res
  commercialRateBps.value = data.commercialRateBps ?? 1200
  runs.value = data.runs ?? []
  if (data.currentPeriodYm) {
    // keep
  }
}

async function loadPeriod(opts?: { syncTab?: boolean }) {
  error.value = ''
  const res: any = await $fetch(`/api/admin/commercial-commissions/periods/${periodYm.value}`)
  const data = res.data ?? res
  lines.value = data.lines ?? []
  totalCents.value = data.totalCents ?? 0
  runStatus.value = data.run?.status || 'open'
  if (opts?.syncTab !== false) {
    if (runStatus.value === 'closed') viewTab.value = 'toPay'
    else if (runStatus.value === 'paid') viewTab.value = 'paid'
    else viewTab.value = 'forecast'
  }
}

async function closePeriod() {
  if (!confirm(t('admin.commercialCommissions.confirmClose', { period: periodYm.value }))) return
  closing.value = true
  error.value = ''
  try {
    await $fetch(`/api/admin/commercial-commissions/periods/${periodYm.value}/close`, { method: 'POST' })
    await loadRunsMeta()
    await loadPeriod()
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    closing.value = false
  }
}

async function markPaid() {
  const missingIban = lines.value.some((l: any) => !l.payoutIban)
  const msg = missingIban
    ? t('admin.commercialCommissions.confirmMarkPaidMissingIban', { period: periodYm.value })
    : t('admin.commercialCommissions.confirmMarkPaid', { period: periodYm.value })
  if (!confirm(msg)) return
  marking.value = true
  error.value = ''
  try {
    await $fetch(`/api/admin/commercial-commissions/periods/${periodYm.value}/mark-paid`, {
      method: 'POST',
      body: { note: '' },
    })
    await loadRunsMeta()
    await loadPeriod()
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    marking.value = false
  }
}

onMounted(async () => {
  await loadRunsMeta()
  await loadPeriod()
})
</script>

<style scoped>
.pro-kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1rem;
  margin-bottom: 1.25rem;
}
.pro-field-inline-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-end;
  margin-bottom: 1rem;
}
.pro-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.pro-tab-count {
  margin-left: 0.25rem;
  opacity: 0.8;
}
.pro-recap {
  margin: 0 0 1rem;
  color: var(--pf-vet-text-muted);
}
.pro-mono {
  font-family: ui-monospace, monospace;
  font-size: 0.85rem;
}
@media (max-width: 900px) {
  .pro-kpi-grid {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
