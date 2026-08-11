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

    <div
      v-if="catalogReady && catalogTotal === 0"
      class="pro-alert pro-mb-md"
      data-testid="medicaments-catalog-empty"
    >
      {{ $t('pharmacy.medicaments.catalogEmpty') }}
      <NuxtLink
        v-if="isAdmin"
        class="pro-link"
        to="/admin/afmps-imports"
        data-testid="medicaments-catalog-import-link"
      >
        {{ $t('pharmacy.medicaments.catalogEmptyLink') }}
      </NuxtLink>
      <span v-else data-testid="medicaments-catalog-empty-hint">
        {{ $t('pharmacy.medicaments.catalogEmptyHint') }}
      </span>
    </div>

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

    <ProCard class="pro-mb-lg" data-testid="medicaments-dictionary">
      <div class="med-dict__head">
        <h3>{{ $t('pharmacy.medicaments.dictionaryTitle') }}</h3>
        <p v-if="catalogTotal > 0" class="pro-hint" data-testid="medicaments-catalog-total">
          {{ $t('pharmacy.medicaments.catalogTotal', { count: catalogTotal }) }}
        </p>
      </div>

      <nav class="med-alpha" :aria-label="$t('pharmacy.medicaments.dictionaryAlphabetAria')" data-testid="medicaments-alphabet">
        <button
          v-for="letter in MEDICAMENT_ALPHABET_LETTERS"
          :key="letter"
          type="button"
          class="med-alpha__btn"
          :class="{ 'is-active': activeLetter === letter }"
          :disabled="!letterHasEntries(letterCounts, letter)"
          :data-testid="`medicaments-letter-${letter === '#' ? 'hash' : letter}`"
          @click="selectLetter(letter)"
        >
          {{ letter }}
        </button>
      </nav>

      <p v-if="dictBusy" class="pro-hint" data-testid="medicaments-dict-loading">
        {{ $t('pharmacy.medicaments.dictionaryLoading') }}
      </p>
      <div v-else-if="!dictItems.length" class="pro-empty" data-testid="medicaments-dict-empty">
        {{ $t('pharmacy.medicaments.dictionaryEmpty') }}
      </div>
      <table v-else class="pro-table" data-testid="medicaments-dict-table">
        <thead>
          <tr>
            <th>{{ $t('pharmacy.medicaments.colName') }}</th>
            <th>{{ $t('pharmacy.medicaments.colCnk') }}</th>
            <th>{{ $t('pharmacy.medicaments.colForm') }}</th>
            <th>{{ $t('pharmacy.medicaments.colAtc') }}</th>
            <th>{{ $t('pharmacy.medicaments.colLink') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in dictItems"
            :key="row.id"
            class="med-dict__row"
            :class="{ 'is-selected': detail?.id === row.id }"
            role="button"
            tabindex="0"
            :data-testid="`medicaments-dict-row-${row.cnk}`"
            @click="openFromDict(row)"
            @keydown.enter.prevent="openFromDict(row)"
            @keydown.space.prevent="openFromDict(row)"
          >
            <td>
              {{ row.name }}
              <ProBadge v-if="row.isAntibiotic" variant="warning" class="med-dict__mini-badge">
                {{ $t('pharmacy.antibioticWarning') }}
              </ProBadge>
            </td>
            <td><code>{{ row.cnk }}</code></td>
            <td>{{ row.pharmaceuticalForm || '—' }}</td>
            <td>{{ row.atcCode || '—' }}</td>
            <td>
              <ProBadge
                v-if="afmpsSourceKind(row.afmpsSource) === 'afmps'"
                variant="success"
                data-testid="medicaments-afmps-badge"
              >
                {{ $t('pharmacy.medicaments.badgeAfmps') }}
              </ProBadge>
              <ProBadge
                v-else-if="afmpsSourceKind(row.afmpsSource) === 'compendium'"
                variant="neutral"
                data-testid="medicaments-compendium-badge"
              >
                {{ $t('pharmacy.medicaments.sourceCompendium') }}
              </ProBadge>
              <ProBadge
                v-else-if="afmpsSourceKind(row.afmpsSource) === 'other'"
                variant="neutral"
                data-testid="medicaments-source-other-badge"
              >
                {{ $t('pharmacy.medicaments.sourceOther') }}
              </ProBadge>
              <span v-else class="pro-hint" data-testid="medicaments-source-unknown">—</span>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="dictTotal > pageSize" class="med-dict__pager" data-testid="medicaments-dict-pager">
        <ProButton
          variant="secondary"
          test-id="medicaments-dict-prev"
          :disabled="dictBusy || dictOffset <= 0"
          @click="pageDict(-1)"
        >
          {{ $t('pharmacy.medicaments.pagePrev') }}
        </ProButton>
        <span class="pro-hint" data-testid="medicaments-dict-page-label">
          {{ $t('pharmacy.medicaments.pageLabel', {
            from: dictOffset + 1,
            to: Math.min(dictOffset + dictItems.length, dictTotal),
            total: dictTotal,
          }) }}
        </span>
        <ProButton
          variant="secondary"
          test-id="medicaments-dict-next"
          :disabled="dictBusy || dictOffset + dictItems.length >= dictTotal"
          @click="pageDict(1)"
        >
          {{ $t('pharmacy.medicaments.pageNext') }}
        </ProButton>
      </div>
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
          <p v-if="metaFields.manufacturer || metaFields.activeSubstance" class="pro-hint" data-testid="medicaments-detail-meta">
            <span v-if="metaFields.manufacturer">{{ metaFields.manufacturer }}</span>
            <span v-if="metaFields.manufacturer && metaFields.activeSubstance"> · </span>
            <span v-if="metaFields.activeSubstance">{{ metaFields.activeSubstance }}</span>
            <span v-if="metaFields.strength"> · {{ metaFields.strength }}</span>
          </p>
        </div>
        <div class="med-detail__badges">
          <ProBadge
            v-if="sourceKind === 'afmps'"
            variant="success"
            data-testid="medicaments-detail-afmps"
          >
            {{ $t('pharmacy.medicaments.badgeAfmps') }}
          </ProBadge>
          <ProBadge
            v-else-if="sourceKind === 'compendium'"
            variant="neutral"
            data-testid="medicaments-detail-source-compendium"
          >
            {{ $t('pharmacy.medicaments.sourceCompendium') }}
          </ProBadge>
          <ProBadge
            v-else-if="sourceKind === 'other'"
            variant="neutral"
            data-testid="medicaments-detail-source-other"
          >
            {{ $t('pharmacy.medicaments.sourceOther') }}
          </ProBadge>
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
import {
  MEDICAMENT_ALPHABET_LETTERS,
  afmpsSourceKind,
  letterHasEntries,
  normalizeMedicamentLetter,
  parseAfmpsMeta,
  parseLetterCounts,
  type MedicamentLetter,
} from '~/utils/medicaments-alphabet'
import { pharmacyErrorMessage } from '~/utils/pharmacy-error'
import { formatPharmacyCents, pharmacyPriceIsPersisted, pharmacyWithdrawalDays } from '~/utils/pharmacy-stock'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pharmacy.read' })

const { t } = useI18n()
const { user } = useProUser()
const { canPractice } = usePracticePerms()
const canWrite = computed(() => canPractice('pharmacy.write'))
const isAdmin = computed(() => user.value?.role === 'admin')

const selected = ref<ProComboboxItem | null>(null)
const detail = ref<MedDetail | null>(null)
const price = ref<MedPrice | null>(null)
const batches = ref<MedBatch[]>([])
const busy = ref(false)
const error = ref('')
const wdMsg = ref('')
const withdrawal = reactive({ meat: 0 as number | null, milk: 0 as number | null, eggs: 0 as number | null, banned: false })

const pageSize = 50
const activeLetter = ref<MedicamentLetter>('A')
const dictItems = ref<MedRow[]>([])
const dictTotal = ref(0)
const dictOffset = ref(0)
const dictBusy = ref(false)
const letterCounts = ref(parseLetterCounts(undefined))
const catalogTotal = ref(0)
const catalogReady = ref(false)
let dictLoadSeq = 0

type MedRow = {
  id: string
  cnk: string
  name: string
  atcCode?: string
  pharmaceuticalForm?: string
  packSize?: string
  isAntibiotic?: boolean
  afmpsSource?: string
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
  afmpsMeta?: unknown
  afmpsSource?: string
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

const metaFields = computed(() => parseAfmpsMeta(detail.value?.afmpsMeta))
const sourceKind = computed(() => afmpsSourceKind(detail.value?.afmpsSource || metaFields.value.source))

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

async function loadDictionary(
  letter: MedicamentLetter,
  offset = 0,
  opts: { includeCounts?: boolean } = {},
) {
  const seq = ++dictLoadSeq
  const includeCounts = opts.includeCounts ?? true
  dictBusy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/medications', {
      query: {
        letter,
        limit: String(pageSize),
        offset: String(offset),
        includeCounts: includeCounts ? '1' : '0',
      },
    })
    if (seq !== dictLoadSeq) return
    const data = unwrap<{
      items?: MedRow[]
      total?: number
      letter?: string
      letterCounts?: Record<string, unknown>
      catalogTotal?: number
    }>(res)
    if (includeCounts && data.letterCounts) {
      letterCounts.value = parseLetterCounts(data.letterCounts)
      catalogTotal.value = data.catalogTotal ?? 0
      catalogReady.value = true
    }

    const resolved = normalizeMedicamentLetter(data.letter) ?? letter
    if (
      includeCounts
      && offset === 0
      && !letterHasEntries(letterCounts.value, resolved)
      && catalogTotal.value > 0
    ) {
      const fallback = MEDICAMENT_ALPHABET_LETTERS.find((l) => letterHasEntries(letterCounts.value, l))
      if (fallback && fallback !== resolved) {
        await loadDictionary(fallback, 0, { includeCounts: true })
        return
      }
    }

    if (seq !== dictLoadSeq) return
    dictItems.value = data.items ?? []
    dictTotal.value = data.total ?? 0
    dictOffset.value = offset
    activeLetter.value = resolved
    if (!catalogReady.value) catalogReady.value = true
  }
  catch (e: any) {
    if (seq !== dictLoadSeq) return
    dictItems.value = []
    error.value = pharmacyErr(e, 'pharmacy.medicaments.errorLoad')
  }
  finally {
    if (seq === dictLoadSeq) dictBusy.value = false
  }
}

function selectLetter(letter: MedicamentLetter) {
  if (!letterHasEntries(letterCounts.value, letter)) return
  loadDictionary(letter, 0, { includeCounts: true })
}

function pageDict(dir: -1 | 1) {
  const next = dictOffset.value + dir * pageSize
  if (next < 0 || next >= dictTotal.value) return
  loadDictionary(activeLetter.value, next, { includeCounts: false })
}

function openFromDict(row: MedRow) {
  selected.value = {
    id: row.id,
    label: row.name,
    hint: row.cnk,
    raw: row,
  }
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

onMounted(() => {
  loadDictionary('A', 0)
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
.med-dict__head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-bottom: 0.75rem;
}
.med-dict__head h3 { margin: 0; }
.med-alpha {
  position: sticky;
  top: 0;
  z-index: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  margin-bottom: 0.85rem;
  padding: 0.35rem 0;
  background: var(--pf-vet-surface);
}
.med-alpha__btn {
  min-width: 1.85rem;
  height: 1.85rem;
  padding: 0 0.35rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 6px;
  background: var(--pf-vet-bg);
  color: var(--pf-vet-primary);
  font: inherit;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
}
.med-alpha__btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.med-alpha__btn.is-active {
  background: var(--pf-vet-accent);
  border-color: var(--pf-vet-accent);
  color: #fff;
}
.med-dict__row { cursor: pointer; }
.med-dict__row:hover { background: color-mix(in srgb, var(--pf-vet-accent) 8%, transparent); }
.med-dict__row.is-selected { background: color-mix(in srgb, var(--pf-vet-accent) 14%, transparent); }
.med-dict__mini-badge { margin-left: 0.35rem; }
.med-dict__pager {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 0.85rem;
}
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
