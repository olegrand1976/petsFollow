<template>
  <div data-testid="medicaments-page">
    <ProPageHeader
      :title="$t('pharmacy.medicaments.title')"
      :subtitle="$t('pharmacy.medicaments.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="medicaments-page-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <details class="pharmacy-legal" data-testid="pharmacy-legal">
      <summary>{{ $t('pharmacy.legal.title') }}</summary>
      <div class="pharmacy-legal__body">
        <p>{{ $t('pharmacy.legal.fefo') }}</p>
        <p>{{ $t('pharmacy.legal.daf') }}</p>
        <p>{{ $t('pharmacy.legal.vamreg') }}</p>
      </div>
    </details>

    <p v-if="error" class="pro-alert pro-mb-md" data-testid="medicaments-error">{{ error }}</p>

    <ProCard class="pro-mb-lg">
      <label class="pro-label" for="med-search">{{ $t('pharmacy.medicaments.searchLabel') }}</label>
      <ProCombobox
        input-id="med-search"
        v-model="selected"
        :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
        :search-fn="searchMedications"
        data-testid="medicaments-search"
      />
    </ProCard>

    <ProCard v-if="detail" class="pro-mb-lg" data-testid="medicaments-detail">
      <div class="med-detail__head">
        <div>
          <h3 data-testid="medicaments-detail-name">{{ detail.name }}</h3>
          <p class="pro-hint">
            CNK {{ detail.cnk }}
            <span v-if="detail.pharmaceuticalForm"> · {{ detail.pharmaceuticalForm }}</span>
            <span v-if="detail.atcCode"> · ATC {{ detail.atcCode }}</span>
            <span v-if="detail.ammNumber"> · AMM {{ detail.ammNumber }}</span>
          </p>
        </div>
        <div class="med-detail__badges">
          <ProBadge v-if="detail.isAntibiotic" variant="warning">{{ $t('pharmacy.antibioticWarning') }}</ProBadge>
          <ProBadge v-if="detail.foodChainBanned" variant="danger">{{ $t('pharmacy.medicaments.foodChainBanned') }}</ProBadge>
        </div>
      </div>

      <div class="med-detail__grid">
        <div>
          <h4>{{ $t('pharmacy.medicaments.withdrawalTitle') }}</h4>
          <div class="stock-form">
            <div>
              <label class="pro-label">{{ $t('pharmacy.medicaments.withdrawalMeat') }}</label>
              <input
                v-model.number="withdrawal.meat"
                type="number"
                min="0"
                class="pro-input"
                data-testid="medicaments-wd-meat"
                :disabled="!canWrite || busy"
              >
            </div>
            <div>
              <label class="pro-label">{{ $t('pharmacy.medicaments.withdrawalMilk') }}</label>
              <input
                v-model.number="withdrawal.milk"
                type="number"
                min="0"
                class="pro-input"
                data-testid="medicaments-wd-milk"
                :disabled="!canWrite || busy"
              >
            </div>
            <div>
              <label class="pro-label">{{ $t('pharmacy.medicaments.withdrawalEggs') }}</label>
              <input
                v-model.number="withdrawal.eggs"
                type="number"
                min="0"
                class="pro-input"
                data-testid="medicaments-wd-eggs"
                :disabled="!canWrite || busy"
              >
            </div>
            <label class="med-detail__check">
              <input
                v-model="withdrawal.banned"
                type="checkbox"
                data-testid="medicaments-wd-banned"
                :disabled="!canWrite || busy"
              >
              {{ $t('pharmacy.medicaments.foodChainBanned') }}
            </label>
            <ProButton
              v-if="canWrite"
              variant="primary"
              test-id="medicaments-wd-save"
              :disabled="busy"
              @click="saveWithdrawal"
            >
              {{ $t('pharmacy.medicaments.withdrawalSave') }}
            </ProButton>
          </div>
          <p v-if="wdMsg" class="pro-hint" data-testid="medicaments-wd-msg">{{ wdMsg }}</p>
        </div>

        <div>
          <h4>{{ $t('pharmacy.medicaments.priceTitle') }}</h4>
          <p v-if="price" data-testid="medicaments-price">
            {{ $t('pharmacy.medicaments.priceSell') }} :
            {{ formatPharmacyCents(price.sellPriceCents) }}
            · {{ $t('pharmacy.medicaments.pricePurchase') }} :
            {{ formatPharmacyCents(price.purchasePriceCents) }}
            · TVA {{ price.vatPercent }}%
          </p>
          <p v-else class="pro-hint" data-testid="medicaments-price-empty">{{ $t('pharmacy.medicaments.priceEmpty') }}</p>
          <NuxtLink class="pro-link" to="/stock" data-testid="medicaments-price-link">
            {{ $t('pharmacy.medicaments.editOnStock') }}
          </NuxtLink>
        </div>
      </div>
    </ProCard>

    <ProCard v-if="detail" data-testid="medicaments-batches">
      <h4 class="pro-mb-sm">{{ $t('pharmacy.medicaments.stockTitle') }}</h4>
      <div v-if="!batches.length" class="pro-empty" data-testid="medicaments-batches-empty">
        {{ $t('pharmacy.medicaments.stockEmpty') }}
      </div>
      <table v-else class="pro-table">
        <thead>
          <tr>
            <th>{{ $t('pharmacy.stock.lot') }}</th>
            <th>{{ $t('pharmacy.stock.deposit') }}</th>
            <th>{{ $t('pharmacy.stock.expiresOn') }}</th>
            <th>{{ $t('pharmacy.stock.qty') }}</th>
            <th>{{ $t('pharmacy.stock.band') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in batches" :key="b.id">
            <td>{{ b.lotNumber }}</td>
            <td>{{ b.depositCode }}</td>
            <td>{{ b.expiresOn }}</td>
            <td>{{ b.qtyOnHand }} {{ b.unit }}</td>
            <td>
              <ProBadge :variant="bandVariant(b.expiryBand)">{{ bandLabel(b.expiryBand) }}</ProBadge>
            </td>
          </tr>
        </tbody>
      </table>
      <p class="pro-hint pro-mt-sm">{{ $t('pharmacy.medicaments.movementsHint') }}</p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import { pharmacyErrorMessage } from '~/utils/pharmacy-error'
import { formatPharmacyCents, pharmacyPriceIsPersisted, pharmacyWithdrawalDays } from '~/utils/pharmacy-stock'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pharmacy.read' })

const { t } = useI18n()
const { canPractice } = usePracticePerms()
const canWrite = computed(() => canPractice('pharmacy.write'))

const selected = ref<ProComboboxItem | null>(null)
const detail = ref<MedDetail | null>(null)
const price = ref<MedPrice | null>(null)
const batches = ref<MedBatch[]>([])
const busy = ref(false)
const error = ref('')
const wdMsg = ref('')
const withdrawal = reactive({ meat: 0 as number | null, milk: 0 as number | null, eggs: 0 as number | null, banned: false })

type MedRow = {
  id: string
  cnk: string
  name: string
  atcCode?: string
  pharmaceuticalForm?: string
  packSize?: string
  isAntibiotic?: boolean
}

type MedDetail = {
  id: string
  cnk: string
  name: string
  atcCode?: string
  pharmaceuticalForm?: string
  packSize?: string
  ammNumber?: string
  isAntibiotic?: boolean
  withdrawalMeatDays?: number | null
  withdrawalMilkDays?: number | null
  withdrawalEggsDays?: number | null
  foodChainBanned?: boolean
}

type MedPrice = {
  purchasePriceCents: number
  sellPriceCents: number
  vatPercent: number
  updatedAt?: string
}

type MedBatch = {
  id: string
  lotNumber: string
  depositCode: string
  expiresOn: string
  qtyOnHand: number
  unit: string
  expiryBand: string
}

function unwrap<T>(res: any): T {
  return (res?.data ?? res) as T
}

function pharmacyErr(e: any, key: string) {
  return pharmacyErrorMessage(t, e, key)
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

async function searchMedications(q: string): Promise<ProComboboxItem[]> {
  const res = await $fetch<any>('/api/vet/pharmacy/medications/search', { query: { q, limit: '20' } })
  const items = unwrap<{ items?: MedRow[] }>(res)?.items ?? []
  return items.map((m) => ({
    id: m.id,
    label: m.name,
    hint: m.cnk + (m.pharmaceuticalForm ? ` · ${m.pharmaceuticalForm}` : ''),
    badge: m.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
    raw: m,
  }))
}

async function loadDetail(id: string) {
  busy.value = true
  error.value = ''
  wdMsg.value = ''
  try {
    const [medRes, priceRes, batchRes] = await Promise.all([
      $fetch<any>(`/api/vet/pharmacy/medications/${id}`),
      $fetch<any>(`/api/vet/pharmacy/prices/${id}`).catch(() => null),
      $fetch<any>('/api/vet/pharmacy/batches', { query: { medicationId: id } }),
    ])
    detail.value = unwrap<MedDetail>(medRes)
    const p = priceRes ? unwrap<MedPrice>(priceRes) : null
    price.value = p && pharmacyPriceIsPersisted(p) ? p : null
    batches.value = unwrap<{ items?: MedBatch[] }>(batchRes)?.items ?? []
    withdrawal.meat = pharmacyWithdrawalDays(detail.value.withdrawalMeatDays)
    withdrawal.milk = pharmacyWithdrawalDays(detail.value.withdrawalMilkDays)
    withdrawal.eggs = pharmacyWithdrawalDays(detail.value.withdrawalEggsDays)
    withdrawal.banned = !!detail.value.foodChainBanned
  }
  catch (e: any) {
    detail.value = null
    error.value = pharmacyErr(e, 'pharmacy.medicaments.errorLoad')
  }
  finally {
    busy.value = false
  }
}

async function saveWithdrawal() {
  if (!detail.value?.id) return
  busy.value = true
  error.value = ''
  wdMsg.value = ''
  try {
    const res = await $fetch<any>(`/api/vet/pharmacy/medications/${detail.value.id}/withdrawal`, {
      method: 'PATCH',
      body: {
        withdrawalMeatDays: withdrawal.meat,
        withdrawalMilkDays: withdrawal.milk,
        withdrawalEggsDays: withdrawal.eggs,
        foodChainBanned: withdrawal.banned,
      },
    })
    detail.value = unwrap<MedDetail>(res)
    wdMsg.value = t('pharmacy.medicaments.withdrawalSaved')
  }
  catch (e: any) {
    error.value = pharmacyErr(e, 'pharmacy.medicaments.errorWithdrawal')
  }
  finally {
    busy.value = false
  }
}

watch(selected, (med) => {
  if (med?.id) {
    loadDetail(med.id)
  }
  else {
    detail.value = null
    price.value = null
    batches.value = []
  }
})
</script>

<style scoped>
.pharmacy-legal {
  margin: 0 0 1rem;
  padding: 0.65rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
}
.pharmacy-legal summary { cursor: pointer; font-weight: 600; }
.pharmacy-legal__body { margin-top: 0.5rem; display: grid; gap: 0.35rem; }
.pharmacy-legal__body p { margin: 0; font-size: 0.9rem; color: var(--pf-vet-muted); }
.med-detail__head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}
.med-detail__badges { display: flex; gap: 0.35rem; flex-wrap: wrap; }
.med-detail__grid {
  display: grid;
  gap: 1.25rem;
  grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr));
}
.stock-form {
  display: grid;
  gap: 0.65rem;
  grid-template-columns: repeat(auto-fill, minmax(8rem, 1fr));
  align-items: end;
}
.med-detail__check {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  grid-column: 1 / -1;
}
.pro-mt-sm { margin-top: 0.5rem; }
.pro-link { color: var(--pf-vet-accent); text-decoration: underline; }
</style>
