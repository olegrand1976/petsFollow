<template>
  <div data-testid="admin-commissions-page">
    <ProPageHeader
      :title="$t('admin.commissions.title')"
      :subtitle="$t('admin.commissions.subtitle')"
    />

    <div class="pro-kpi-grid">
      <ProKpi :label="$t('admin.commissions.kpiPeriod')" :value="periodYm" />
      <ProKpi :label="$t('admin.commissions.kpiDue')" :value="formatCurrency(totalCents)" />
      <ProKpi :label="$t('admin.commissions.kpiStatus')" :value="runStatusLabel" />
    </div>

    <ProCard :title="$t('admin.commissions.tiersTitle')" class="pro-mb">
      <ul class="pro-commissions-tiers">
        <li v-for="t in tiers" :key="t.minClients">
          {{ formatTier(t) }}
        </li>
      </ul>
    </ProCard>

    <ProCard :title="$t('admin.commissions.periodTitle')">
      <div class="pro-list-toolbar__filters pro-field-inline-wrap">
        <div class="pro-field pro-field-inline">
          <label class="pro-label" for="period-ym">{{ $t('admin.commissions.period') }}</label>
          <input id="period-ym" v-model="periodYm" class="pro-input" type="month" @change="loadPeriod">
        </div>
        <ProButton
          variant="secondary"
          :disabled="runStatus !== 'open' || closing"
          :loading="closing"
          test-id="commissions-close"
          @click="closePeriod"
        >
          {{ $t('admin.commissions.close') }}
        </ProButton>
        <ProButton
          :disabled="runStatus !== 'closed' || marking"
          :loading="marking"
          test-id="commissions-mark-paid"
          @click="markPaid"
        >
          {{ $t('admin.commissions.markPaid') }}
        </ProButton>
      </div>

      <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

      <ProTable :empty="!lines.length" :empty-title="$t('admin.commissions.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commissions.colVet') }}</th>
            <th>{{ $t('admin.commissions.colClients') }}</th>
            <th>{{ $t('admin.commissions.colLedger') }}</th>
            <th>{{ $t('admin.commissions.colAmount') }}</th>
            <th>{{ $t('admin.commissions.colStatus') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="l in lines" :key="l.vetUserId + (l.id || '')">
            <td>
              <div>{{ l.vetFullName }}</div>
              <div class="text-muted">{{ l.vetEmail }}</div>
            </td>
            <td>{{ l.eligibleClients }}</td>
            <td>{{ l.ledgerCount }}</td>
            <td>{{ formatCurrency(l.amountCents) }}</td>
            <td>
              <ProBadge :variant="lineVariant(l.status || runStatus)">
                {{ l.status || runStatus }}
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

const now = new Date()
const periodYm = ref(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`)
const lines = ref<any[]>([])
const totalCents = ref(0)
const runStatus = ref('open')
const tiers = ref<any[]>([])
const closing = ref(false)
const marking = ref(false)
const error = ref('')

const runStatusLabel = computed(() => {
  switch (runStatus.value) {
    case 'open':
      return t('admin.commissions.statusOpen')
    case 'closed':
      return t('admin.commissions.statusClosed')
    case 'paid':
      return t('admin.commissions.statusPaid')
    default:
      return runStatus.value
  }
})

function formatTier(tier: any) {
  const max = tier.maxClients == null ? '+' : `–${tier.maxClients}`
  const pct = (tier.rateBps / 100).toFixed(0)
  return t('admin.commissions.tierLine', {
    min: tier.minClients,
    max,
    pct,
  })
}

function lineVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  if (status === 'paid') return 'success'
  if (status === 'closed' || status === 'pending') return 'warning'
  return 'neutral'
}

async function loadRunsMeta() {
  const res: any = await $fetch('/api/admin/commissions/runs')
  const data = res.data ?? res
  tiers.value = data.tiers ?? []
}

async function loadPeriod() {
  error.value = ''
  const res: any = await $fetch(`/api/admin/commissions/periods/${periodYm.value}`)
  const data = res.data ?? res
  lines.value = data.lines ?? []
  totalCents.value = data.totalCents ?? 0
  runStatus.value = data.run?.status || 'open'
}

async function closePeriod() {
  closing.value = true
  error.value = ''
  try {
    await $fetch(`/api/admin/commissions/periods/${periodYm.value}/close`, { method: 'POST' })
    await loadPeriod()
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    closing.value = false
  }
}

async function markPaid() {
  marking.value = true
  error.value = ''
  try {
    await $fetch(`/api/admin/commissions/periods/${periodYm.value}/mark-paid`, {
      method: 'POST',
      body: { note: '' },
    })
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
  grid-template-columns: repeat(3, minmax(0, 1fr));
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

.pro-field-inline-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-end;
  margin-bottom: 1rem;
}

@media (max-width: 768px) {
  .pro-kpi-grid {
    grid-template-columns: 1fr;
  }
}
</style>
