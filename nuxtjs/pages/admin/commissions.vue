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

    <ProCard :title="$t('admin.commissions.sheetTitle')" class="pro-mb">
      <ProCommissionSheet
        audience="admin"
        :plan-rates="planRates"
        :addon-rates="addonRates"
        :bonuses="bonuses"
      />
    </ProCard>

    <ProCard :title="$t('admin.commissions.tiersTitle')" class="pro-mb">
      <p class="text-muted pro-mb-sm">{{ $t('admin.commissions.planRatesHint') }}</p>
      <div v-for="(tier, idx) in editTiers" :key="idx" class="pro-tier-row">
        <input v-model.number="tier.minClients" class="pro-input pro-input-narrow" type="number" min="1">
        <input
          v-model.number="tier.maxClients"
          class="pro-input pro-input-narrow"
          type="number"
          min="1"
          :placeholder="$t('admin.commissions.openEnded')"
        >
        <input v-model.number="tier.ratePct" class="pro-input pro-input-narrow" type="number" min="0" max="50">
        <span>%</span>
      </div>
      <p class="text-muted">{{ $t('admin.commissions.tiersHint') }}</p>
      <ProButton :loading="savingTiers" @click="saveTiers">
        {{ $t('admin.commissions.saveTiers') }}
      </ProButton>
      <p v-if="settingsError" class="pro-field-error" role="alert">{{ settingsError }}</p>
    </ProCard>

    <ProCard :title="$t('admin.commissions.profileRatesTitle')" class="pro-mb" data-testid="admin-profile-rates">
      <p class="text-muted pro-mb-sm">{{ $t('admin.commissions.profileRatesHint') }}</p>
      <p v-if="!profileLedgerWired" class="pro-hint pro-mb-sm" data-testid="admin-profile-rates-stub">
        {{ $t('admin.commissions.profileRatesLedgerStub') }}
      </p>
      <ProTable v-if="profileRates.length">
        <thead>
          <tr>
            <th>{{ $t('admin.commissions.profileKey') }}</th>
            <th>{{ $t('admin.commissions.profileLabel') }}</th>
            <th>{{ $t('admin.commissions.profileRate') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in profileRates" :key="row.profileKey">
            <td class="pro-mono">{{ row.profileKey }}</td>
            <td>{{ row.label }}</td>
            <td>
              <input
                v-model.number="row.ratePct"
                class="pro-input pro-input-narrow"
                type="number"
                min="0"
                max="100"
                step="0.01"
              >
              <span>%</span>
            </td>
          </tr>
        </tbody>
      </ProTable>
      <ProButton class="pro-mt-md" :loading="savingProfileRates" test-id="save-profile-rates" @click="saveProfileRates">
        {{ $t('admin.commissions.saveProfileRates') }}
      </ProButton>
      <p v-if="profileRatesError" class="pro-field-error" role="alert">{{ profileRatesError }}</p>
      <div v-if="commercialPlanRates" class="pro-mt-md text-muted">
        <p>{{ $t('admin.commissions.commercialPlanRatesReadonly') }}</p>
        <ul>
          <li>monthly: {{ (commercialPlanRates.monthly / 100).toFixed(0) }} %</li>
          <li>annual: {{ (commercialPlanRates.annual / 100).toFixed(0) }} %</li>
          <li>triennial: {{ (commercialPlanRates.triennial / 100).toFixed(0) }} %</li>
        </ul>
      </div>
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
          :disabled="!canMarkReady || marking"
          :loading="marking"
          test-id="commissions-mark-paid"
          @click="markPaid"
        >
          {{ $t('admin.commissions.markReadyPaid') }}
        </ProButton>
      </div>

      <p v-if="error" class="pro-field-error" role="alert">{{ error }}</p>

      <ProTable :empty="!lines.length" :empty-title="$t('admin.commissions.empty')">
        <thead>
          <tr>
            <th>{{ $t('admin.commissions.colVet') }}</th>
            <th>{{ $t('admin.commissions.colCompany') }}</th>
            <th>{{ $t('admin.commissions.colIban') }}</th>
            <th>{{ $t('admin.commissions.colClients') }}</th>
            <th>{{ $t('admin.commissions.colLedger') }}</th>
            <th>{{ $t('admin.commissions.colAmount') }}</th>
            <th>{{ $t('admin.commissions.colStatus') }}</th>
            <th>{{ $t('admin.commissions.colActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="l in lines" :key="l.vetUserId + (l.id || '')">
            <td>
              <div>{{ l.vetFullName }}</div>
              <div class="text-muted">{{ l.vetEmail }}</div>
            </td>
            <td>
              <div>{{ l.companyLegalName || '—' }}</div>
              <div v-if="l.vatNumber" class="text-muted">{{ l.vatNumber }}</div>
            </td>
            <td>
              <div class="pro-mono">{{ l.payoutIban || '—' }}</div>
              <div v-if="l.payoutAccountHolder" class="text-muted">{{ l.payoutAccountHolder }}</div>
            </td>
            <td>{{ l.eligibleClients }}</td>
            <td>{{ l.ledgerCount }}</td>
            <td>{{ formatCurrency(l.amountCents) }}</td>
            <td>
              <ProBadge :variant="lineVariant(l.status || runStatus)">
                {{ lineStatusLabel(l.status || runStatus) }}
              </ProBadge>
            </td>
            <td>
              <ProButton
                v-if="l.status === 'ready_to_pay'"
                variant="secondary"
                :loading="markingLineId === l.vetUserId"
                @click="markLinePaid(l.vetUserId)"
              >
                {{ $t('admin.commissions.markLinePaid') }}
              </ProButton>
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
const editTiers = ref<{ minClients: number, maxClients: number | null, ratePct: number }[]>([])
const planRates = ref<any[]>([])
const addonRates = ref<any[]>([])
const bonuses = ref<any[]>([])
const closing = ref(false)
const marking = ref(false)
const markingLineId = ref('')
const savingTiers = ref(false)
const error = ref('')
const settingsError = ref('')
const profileRates = ref<{ profileKey: string, label: string, rateBps: number, ratePct: number }[]>([])
const commercialPlanRates = ref<Record<string, number> | null>(null)
const profileLedgerWired = ref(true)
const savingProfileRates = ref(false)
const profileRatesError = ref('')

const runStatusLabel = computed(() => {
  switch (runStatus.value) {
    case 'open':
      return t('admin.commissions.statusOpen')
    case 'closed':
      return t('admin.commissions.statusClosed')
    case 'partially_paid':
      return t('admin.commissions.statusPartiallyPaid')
    case 'paid':
      return t('admin.commissions.statusPaid')
    default:
      return runStatus.value
  }
})

const canMarkReady = computed(() =>
  (runStatus.value === 'closed' || runStatus.value === 'partially_paid')
  && lines.value.some((l: any) => l.status === 'ready_to_pay'),
)

function syncEditTiers(list: any[]) {
  editTiers.value = (list || []).map((tier: any) => ({
    minClients: tier.minClients,
    maxClients: tier.maxClients ?? null,
    ratePct: Math.round((tier.rateBps || 0) / 100),
  }))
}

async function saveTiers() {
  savingTiers.value = true
  settingsError.value = ''
  try {
    const tiersPayload = editTiers.value.map((tier) => ({
      minClients: tier.minClients,
      maxClients: tier.maxClients == null || Number.isNaN(tier.maxClients as number) ? null : tier.maxClients,
      rateBps: Math.round(tier.ratePct * 100),
    }))
    const res: any = await $fetch('/api/admin/commissions/tiers', {
      method: 'PUT',
      body: { tiers: tiersPayload },
    })
    const data = res.data ?? res
    tiers.value = data.tiers ?? []
    syncEditTiers(tiers.value)
  } catch (e: any) {
    settingsError.value = mapError(e)
  } finally {
    savingTiers.value = false
  }
}

async function loadProfileRates() {
  profileRatesError.value = ''
  try {
    const res: any = await $fetch('/api/admin/commissions/profile-rates')
    const data = res.data ?? res
    profileRates.value = (data.profileRates || []).map((r: any) => ({
      profileKey: r.profileKey,
      label: r.label,
      rateBps: r.rateBps || 0,
      ratePct: (r.rateBps || 0) / 100,
    }))
    commercialPlanRates.value = data.commercialPlanRates || null
    profileLedgerWired.value = data.ledgerWired === true
  } catch (e: any) {
    profileRatesError.value = mapError(e)
  }
}

async function saveProfileRates() {
  savingProfileRates.value = true
  profileRatesError.value = ''
  try {
    const res: any = await $fetch('/api/admin/commissions/profile-rates', {
      method: 'PUT',
      body: {
        rates: profileRates.value.map((r) => ({
          profileKey: r.profileKey,
          label: r.label,
          rateBps: Math.round(r.ratePct * 100),
        })),
      },
    })
    const data = res.data ?? res
    profileRates.value = (data.profileRates || []).map((r: any) => ({
      profileKey: r.profileKey,
      label: r.label,
      rateBps: r.rateBps || 0,
      ratePct: (r.rateBps || 0) / 100,
    }))
  } catch (e: any) {
    profileRatesError.value = mapError(e)
  } finally {
    savingProfileRates.value = false
  }
}

function lineStatusLabel(status: string) {
  const key = `admin.commissions.lineStatus.${status}`
  const label = t(key)
  if (label !== key) return label
  switch (status) {
    case 'open': return t('admin.commissions.statusOpen')
    case 'closed': return t('admin.commissions.statusClosed')
    case 'partially_paid': return t('admin.commissions.statusPartiallyPaid')
    case 'paid': return t('admin.commissions.statusPaid')
    default: return status
  }
}

function lineVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'paid':
      return 'success'
    case 'ready_to_pay':
    case 'closed':
      return 'warning'
    case 'missing_info':
      return 'danger'
    default:
      return 'neutral'
  }
}

async function loadRunsMeta() {
  const res: any = await $fetch('/api/admin/commissions/runs')
  const data = res.data ?? res
  tiers.value = data.tiers ?? []
  syncEditTiers(tiers.value)
  planRates.value = data.planRates ?? []
  addonRates.value = data.addonRates ?? []
  bonuses.value = data.bonuses ?? []
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
  if (!confirm(t('admin.commissions.confirmClose', { period: periodYm.value }))) return
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
  const missing = lines.value.filter((l: any) => l.status === 'missing_info').length
  const msg = missing
    ? t('admin.commissions.confirmMarkReadyMissing', { period: periodYm.value, n: missing })
    : t('admin.commissions.confirmMarkReady', { period: periodYm.value })
  if (!confirm(msg)) return
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

async function markLinePaid(vetUserId: string) {
  if (!confirm(t('admin.commissions.confirmMarkLinePaid'))) return
  markingLineId.value = vetUserId
  error.value = ''
  try {
    await $fetch(`/api/admin/commissions/periods/${periodYm.value}/lines/${vetUserId}/mark-paid`, {
      method: 'POST',
    })
    await loadPeriod()
  } catch (e: any) {
    error.value = mapError(e)
  } finally {
    markingLineId.value = ''
  }
}

onMounted(async () => {
  await loadRunsMeta()
  await loadPeriod()
  await loadProfileRates()
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

.pro-tier-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;
}
.pro-input-narrow {
  max-width: 6rem;
}
.pro-mb-sm {
  margin-bottom: 1rem;
}

.pro-field-inline-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-end;
  margin-bottom: 1rem;
}

.pro-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.85rem;
}

@media (max-width: 768px) {
  .pro-kpi-grid {
    grid-template-columns: 1fr;
  }
}
</style>
