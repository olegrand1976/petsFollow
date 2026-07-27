<template>
  <div data-testid="stock-page">
    <ProPageHeader
      :title="$t('pharmacy.stock.title')"
      :subtitle="$t('pharmacy.stock.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="stock-page-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton variant="secondary" test-id="stock-export" @click="exportCsv">
          {{ $t('pharmacy.stock.export') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <details class="pharmacy-legal" data-testid="pharmacy-legal">
      <summary>{{ $t('pharmacy.legal.title') }}</summary>
      <div class="pharmacy-legal__body">
        <p>{{ $t('pharmacy.legal.fefo') }}</p>
        <p>{{ $t('pharmacy.legal.expiry') }}</p>
        <p>{{ $t('pharmacy.legal.waste') }}</p>
      </div>
    </details>

    <p v-if="error" class="pro-alert pro-mb-md" data-testid="stock-error">{{ error }}</p>

    <div class="stock-bands" data-testid="stock-bands">
      <button
        v-for="b in bandCards"
        :key="b.key"
        type="button"
        class="stock-band"
        :class="{ 'stock-band--active': bandFilter === b.key }"
        :data-testid="`stock-band-${b.key}`"
        @click="setBand(b.key)"
      >
        <span class="stock-band__label">{{ b.label }}</span>
        <strong class="stock-band__n">{{ summary[b.field] ?? 0 }}</strong>
      </button>
    </div>

    <ProCard class="pro-mb-lg" data-testid="stock-receipt">
      <h3 class="pro-mb-sm">{{ $t('pharmacy.stock.receiptTitle') }}</h3>
      <div class="stock-form">
        <div>
          <label class="pro-label" for="stock-med">{{ $t('pharmacy.stock.medication') }}</label>
          <ProCombobox
            input-id="stock-med"
            v-model="selectedMed"
            :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
            :search-fn="searchMedications"
            data-testid="stock-med-search"
          />
        </div>
        <div>
          <label class="pro-label" for="stock-lot">{{ $t('pharmacy.stock.lot') }}</label>
          <input id="stock-lot" v-model="receipt.lotNumber" class="pro-input" data-testid="stock-lot" />
        </div>
        <div>
          <label class="pro-label" for="stock-exp">{{ $t('pharmacy.stock.expiresOn') }}</label>
          <input id="stock-exp" v-model="receipt.expiresOn" type="date" class="pro-input" data-testid="stock-exp" />
        </div>
        <div>
          <label class="pro-label" for="stock-qty">{{ $t('pharmacy.stock.qty') }}</label>
          <input id="stock-qty" v-model.number="receipt.qty" type="number" min="0.001" step="any" class="pro-input" data-testid="stock-qty" />
        </div>
        <div class="stock-form__actions">
          <ProButton variant="primary" test-id="stock-receive" :disabled="busy || !canReceive" @click="receive">
            {{ $t('pharmacy.stock.receive') }}
          </ProButton>
        </div>
      </div>
      <p v-if="softWarn" class="pro-hint" data-testid="stock-soft-warn">{{ $t('pharmacy.stock.shortDatedWarn') }}</p>
    </ProCard>

    <ProCard data-testid="stock-batches">
      <div class="stock-toolbar">
        <input
          v-model="q"
          class="pro-input"
          :placeholder="$t('pharmacy.stock.searchPlaceholder')"
          data-testid="stock-search"
          @keyup.enter="loadBatches"
        />
        <ProButton variant="secondary" test-id="stock-refresh" :disabled="busy" @click="refresh">
          {{ $t('pharmacy.stock.refresh') }}
        </ProButton>
      </div>
      <div v-if="!batches.length" class="pro-empty" data-testid="stock-empty">
        {{ $t('pharmacy.stock.empty') }}
      </div>
      <table v-else class="pro-table" data-testid="stock-table">
        <thead>
          <tr>
            <th>{{ $t('pharmacy.stock.colMed') }}</th>
            <th>{{ $t('pharmacy.stock.lot') }}</th>
            <th>{{ $t('pharmacy.stock.expiresOn') }}</th>
            <th>{{ $t('pharmacy.stock.qty') }}</th>
            <th>{{ $t('pharmacy.stock.band') }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in batches" :key="row.id" :data-testid="`stock-row-${row.id}`">
            <td>
              <strong>{{ row.medicationName }}</strong>
              <div class="pro-hint">{{ row.medicationCnk }} · {{ row.depositCode }}</div>
            </td>
            <td>{{ row.lotNumber }}</td>
            <td>{{ row.expiresOn }}</td>
            <td>{{ row.qtyOnHand }} {{ row.unit }}</td>
            <td>
              <ProBadge :variant="bandVariant(row.expiryBand)">{{ bandLabel(row.expiryBand) }}</ProBadge>
            </td>
            <td class="stock-row-actions">
              <ProButton
                v-if="row.status === 'active'"
                variant="secondary"
                :test-id="`stock-quarantine-${row.id}`"
                :disabled="busy"
                @click="quarantine(row.id)"
              >
                {{ $t('pharmacy.stock.quarantine') }}
              </ProButton>
              <ProButton
                variant="ghost"
                :test-id="`stock-waste-${row.id}`"
                :disabled="busy"
                @click="waste(row.id)"
              >
                {{ $t('pharmacy.stock.waste') }}
              </ProButton>
            </td>
          </tr>
        </tbody>
      </table>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'

definePageMeta({ middleware: ['auth', 'vet-only'] })

const { t } = useI18n()
const busy = ref(false)
const error = ref('')
const softWarn = ref(false)
const bandFilter = ref('all')
const q = ref('')
const selectedMed = ref<ProComboboxItem | null>(null)
const receipt = reactive({ lotNumber: '', expiresOn: '', qty: 1 })
const batches = ref<BatchRow[]>([])
const summary = ref<Record<string, number>>({ ok: 0, soon: 0, return: 0, critical: 0, expired: 0, quarantine: 0 })

type BatchRow = {
  id: string
  medicationName: string
  medicationCnk: string
  depositCode: string
  lotNumber: string
  expiresOn: string
  qtyOnHand: number
  unit: string
  status: string
  expiryBand: string
}

type MedRow = {
  id: string
  cnk: string
  name: string
  pharmaceuticalForm?: string
  isAntibiotic?: boolean
}

const bandCards = computed(() => [
  { key: 'all', field: '_all', label: t('pharmacy.stock.bandAll') },
  { key: 'critical', field: 'critical', label: t('pharmacy.stock.bandCritical') },
  { key: 'return', field: 'return', label: t('pharmacy.stock.bandReturn') },
  { key: 'soon', field: 'soon', label: t('pharmacy.stock.bandSoon') },
  { key: 'expired', field: 'expired', label: t('pharmacy.stock.bandExpired') },
  { key: 'quarantine', field: 'quarantine', label: t('pharmacy.stock.bandQuarantine') },
])

const canReceive = computed(() =>
  Boolean(selectedMed.value?.id && receipt.lotNumber && receipt.expiresOn && receipt.qty > 0),
)

function unwrapData<T>(res: any): T {
  return (res?.data ?? res) as T
}

function bandVariant(band: string): 'success' | 'warning' | 'danger' | 'neutral' {
  if (band === 'ok') return 'success'
  if (band === 'soon' || band === 'return') return 'warning'
  if (band === 'critical' || band === 'expired' || band === 'quarantine') return 'danger'
  return 'neutral'
}

function bandLabel(band: string) {
  const map: Record<string, string> = {
    ok: t('pharmacy.stock.bandOk'),
    soon: t('pharmacy.stock.bandSoon'),
    return: t('pharmacy.stock.bandReturn'),
    critical: t('pharmacy.stock.bandCritical'),
    expired: t('pharmacy.stock.bandExpired'),
    quarantine: t('pharmacy.stock.bandQuarantine'),
  }
  return map[band] || band
}

function setBand(key: string) {
  bandFilter.value = key
  loadBatches()
}

async function searchMedications(query: string): Promise<ProComboboxItem[]> {
  const res = await $fetch<any>('/api/vet/pharmacy/medications/search', { query: { q: query, limit: '20' } })
  const items = unwrapData<{ items?: MedRow[] }>(res)?.items ?? []
  return items.map((m) => ({
    id: m.id,
    label: m.name,
    hint: m.cnk + (m.pharmaceuticalForm ? ` · ${m.pharmaceuticalForm}` : ''),
    badge: m.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
    raw: m,
  }))
}

async function loadSummary() {
  const res = await $fetch<any>('/api/vet/pharmacy/expiry/summary')
  const data = unwrapData<Record<string, number>>(res)
  summary.value = {
    ok: data.ok ?? 0,
    soon: data.soon ?? 0,
    return: data.return ?? 0,
    critical: data.critical ?? 0,
    expired: data.expired ?? 0,
    quarantine: data.quarantine ?? 0,
    _all: (data.ok ?? 0) + (data.soon ?? 0) + (data.return ?? 0) + (data.critical ?? 0) + (data.expired ?? 0) + (data.quarantine ?? 0),
  }
}

async function loadBatches() {
  const res = await $fetch<any>('/api/vet/pharmacy/batches', {
    query: {
      band: bandFilter.value === 'all' ? undefined : bandFilter.value,
      q: q.value || undefined,
    },
  })
  batches.value = unwrapData<{ items?: BatchRow[] }>(res)?.items ?? []
}

async function refresh() {
  busy.value = true
  error.value = ''
  try {
    await Promise.all([loadSummary(), loadBatches()])
  }
  catch (e: any) {
    error.value = e?.data?.error?.message || e?.message || t('pharmacy.stock.errorLoad')
  }
  finally {
    busy.value = false
  }
}

async function receive() {
  if (!canReceive.value || !selectedMed.value) return
  busy.value = true
  error.value = ''
  softWarn.value = false
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/batches', {
      method: 'POST',
      body: {
        medicationId: selectedMed.value.id,
        lotNumber: receipt.lotNumber,
        expiresOn: receipt.expiresOn,
        qty: receipt.qty,
      },
    })
    const data = unwrapData<{ shortDatedWarning?: boolean }>(res)
    softWarn.value = Boolean(data.shortDatedWarning)
    receipt.lotNumber = ''
    receipt.qty = 1
    await refresh()
  }
  catch (e: any) {
    error.value = e?.data?.error?.code || e?.data?.error?.message || t('pharmacy.stock.errorReceive')
  }
  finally {
    busy.value = false
  }
}

async function quarantine(id: string) {
  busy.value = true
  error.value = ''
  try {
    await $fetch(`/api/vet/pharmacy/batches/${id}/quarantine`, { method: 'POST', body: { reason: 'manual' } })
    await refresh()
  }
  catch (e: any) {
    error.value = e?.data?.error?.message || t('pharmacy.stock.errorAction')
  }
  finally {
    busy.value = false
  }
}

async function waste(id: string) {
  if (!confirm(t('pharmacy.stock.wasteConfirm'))) return
  busy.value = true
  error.value = ''
  try {
    await $fetch(`/api/vet/pharmacy/batches/${id}/waste`, { method: 'POST', body: { reason: 'expired' } })
    await refresh()
  }
  catch (e: any) {
    error.value = e?.data?.error?.message || t('pharmacy.stock.errorAction')
  }
  finally {
    busy.value = false
  }
}

function exportCsv() {
  window.open('/api/vet/pharmacy/batches/export.csv', '_blank')
}

onMounted(refresh)
</script>

<style scoped>
.pharmacy-legal {
  margin: 0 0 1rem;
  padding: 0.65rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
}
.pharmacy-legal summary {
  cursor: pointer;
  font-weight: 600;
}
.pharmacy-legal__body {
  margin-top: 0.5rem;
  display: grid;
  gap: 0.35rem;
}
.pharmacy-legal__body p { margin: 0; font-size: 0.9rem; color: var(--pf-vet-muted); }
.stock-bands {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(7.5rem, 1fr));
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.stock-band {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
  padding: 0.65rem 0.75rem;
  text-align: left;
  cursor: pointer;
}
.stock-band--active { outline: 2px solid var(--pf-vet-primary); }
.stock-band__label { display: block; font-size: 0.75rem; color: var(--pf-vet-muted); }
.stock-band__n { font-size: 1.25rem; }
.stock-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.75rem;
  align-items: end;
}
.stock-form__actions { display: flex; align-items: end; }
.stock-toolbar {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}
.stock-row-actions {
  display: flex;
  gap: 0.35rem;
  justify-content: flex-end;
  flex-wrap: wrap;
}
</style>
