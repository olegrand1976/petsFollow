<template>
  <div data-testid="daf-wizard-page">
    <ProPageHeader
      :title="$t('pharmacy.daf.wizardTitle')"
      :subtitle="$t('pharmacy.daf.wizardSubtitle')"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <p v-if="error" class="pro-alert" data-testid="daf-wizard-error">{{ error }}</p>

    <ProCard v-if="contextLabel || petRegulatoryLabel || visitId" class="pro-mb-lg" data-testid="daf-consultation-context">
      <p v-if="visitId" class="daf-from-consult" data-testid="daf-from-consultation-banner">
        <strong>{{ $t('pharmacy.daf.fromConsultation') }}</strong>
        <NuxtLink to="/consultations" class="pro-link" data-testid="daf-back-consultations">
          {{ $t('pharmacy.daf.backToConsultations') }}
        </NuxtLink>
      </p>
      <p v-if="contextLabel" class="pro-hint">{{ contextLabel }}</p>
      <dl v-if="petRegulatoryLabel" class="daf-pet-reg" data-testid="daf-pet-regulatory">
        <div>
          <dt>{{ $t('clients.pet.foodChainStatus') }}</dt>
          <dd>{{ petFoodChainLabel }}</dd>
        </div>
        <div v-if="petDomicile">
          <dt>{{ $t('clients.pet.domicileLocation') }}</dt>
          <dd>{{ petDomicile }}</dd>
        </div>
      </dl>
    </ProCard>

    <ProCard class="pro-mb-lg">
      <div v-for="(line, idx) in lines" :key="idx" class="daf-line" :data-testid="`daf-line-${idx}`">
        <div>
          <label class="pro-label">{{ $t('pharmacy.stock.medication') }}</label>
          <ProCombobox
            :input-id="`daf-med-${idx}`"
            v-model="line.med"
            :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
            :search-fn="searchMedications"
          />
        </div>
        <div>
          <label class="pro-label">{{ $t('pharmacy.daf.amm') }}</label>
          <input v-model="line.ammNumber" class="pro-input" :data-testid="`daf-amm-${idx}`" />
        </div>
        <div>
          <label class="pro-label">{{ $t('pharmacy.stock.qty') }}</label>
          <input v-model.number="line.qty" type="number" min="0.001" step="any" class="pro-input" />
        </div>
        <template v-if="line.med?.raw?.isAntibiotic">
          <div>
            <label class="pro-label">{{ $t('pharmacy.daf.vamregSpecies') }}</label>
            <input v-model="line.species" class="pro-input" />
          </div>
          <div>
            <label class="pro-label">{{ $t('pharmacy.daf.vamregIndication') }}</label>
            <input v-model="line.indication" class="pro-input" />
          </div>
          <div>
            <label class="pro-label">{{ $t('pharmacy.daf.vamregDuration') }}</label>
            <input v-model.number="line.durationDays" type="number" min="1" class="pro-input" />
          </div>
        </template>
      </div>
      <ProButton variant="secondary" test-id="daf-add-line" @click="addLine">{{ $t('pharmacy.daf.addLine') }}</ProButton>
    </ProCard>

    <ProCard v-if="preview.length" class="pro-mb-lg" data-testid="daf-preview">
      <h3>{{ $t('pharmacy.daf.previewTitle') }}</h3>
      <ul>
        <li v-for="(p, i) in preview" :key="i">
          {{ p.medicationName }} · lot {{ p.lotNumber }} · {{ p.qty }} · DLC {{ p.expiresOn }}
        </li>
      </ul>
    </ProCard>

    <div class="daf-actions">
      <ProButton variant="secondary" test-id="daf-preview-btn" :disabled="busy" @click="runPreview">{{ $t('pharmacy.daf.preview') }}</ProButton>
      <ProButton variant="secondary" test-id="daf-save-draft" :disabled="busy || !canSubmit" @click="saveDraft">{{ $t('pharmacy.daf.saveDraft') }}</ProButton>
      <ProButton variant="primary" test-id="daf-finalize-btn" :disabled="busy || !canSubmit" @click="requestFinalize">{{ $t('pharmacy.daf.finalize') }}</ProButton>
      <ProButton
        v-if="finalizedDafId && canInvoice"
        test-id="daf-go-invoice"
        :disabled="busy"
        @click="goInvoice"
      >
        {{ $t('pharmacy.daf.goInvoice') }}
      </ProButton>
    </div>

    <ProCard
      v-if="stockReceiptMed"
      class="pro-mb-lg"
      data-testid="daf-receipt-express"
    >
      <h3>{{ $t('clients.consultation.receiptExpressTitle') }}</h3>
      <p class="pro-hint">{{ stockReceiptMed.name }}</p>
      <div class="daf-receipt-form">
        <div>
          <label class="pro-label">{{ $t('clients.consultation.receiptLot') }}</label>
          <input v-model="receiptForm.lotNumber" class="pro-input" data-testid="daf-receipt-lot" :disabled="busy" />
        </div>
        <div>
          <label class="pro-label">{{ $t('clients.consultation.receiptExpiry') }}</label>
          <input v-model="receiptForm.expiresOn" type="date" class="pro-input" data-testid="daf-receipt-expiry" :disabled="busy" />
        </div>
        <div>
          <label class="pro-label">{{ $t('clients.consultation.receiptQty') }}</label>
          <input v-model.number="receiptForm.qty" type="number" min="0.001" step="any" class="pro-input" data-testid="daf-receipt-qty" :disabled="busy" />
        </div>
      </div>
      <ProButton
        variant="primary"
        test-id="daf-receipt-submit"
        :disabled="busy || !receiptForm.lotNumber.trim() || !receiptForm.expiresOn || receiptForm.qty <= 0"
        @click="submitReceipt"
      >
        {{ $t('clients.consultation.receiptSubmit') }}
      </ProButton>
    </ProCard>

    <ProModal
      :open="finalizeConfirmOpen"
      size="md"
      :title="$t('clients.consultation.finalizeConfirmTitle')"
      test-id="daf-finalize-confirm"
      :prevent-close="busy"
      @update:open="onFinalizeConfirmOpen"
    >
      <p class="pro-hint">{{ $t('clients.consultation.finalizeConfirmBody') }}</p>
      <template #footer>
        <ProButton variant="ghost" test-id="daf-finalize-cancel" :disabled="busy" @click="finalizeConfirmOpen = false">
          {{ $t('common.cancel') }}
        </ProButton>
        <ProButton variant="primary" test-id="daf-finalize-ok" :loading="busy" @click="confirmFinalize">
          {{ $t('clients.consultation.finalizeConfirmOk') }}
        </ProButton>
      </template>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import { INVOICING_UI_ENABLED } from '~/utils/invoicing-ui'
import { pharmacyErrorMessage } from '~/utils/pharmacy-error'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pharmacy.write' })
const { t } = useI18n()
function pharmacyErr(e: any, fallbackKey: string): string {
  return pharmacyErrorMessage(t, e, fallbackKey)
}
const route = useRoute()
const runtimeConfig = useRuntimeConfig()
const { canPractice } = usePracticePerms()
// Sans Billit actif la page /invoicing redirige : ne pas proposer le CTA.
const canInvoice = computed(() => (
  INVOICING_UI_ENABLED
  && isPublicFlagOn(runtimeConfig.public.billitEnabled)
  && canPractice('clients.write')
))
const busy = ref(false)
const error = ref('')
const preview = ref<any[]>([])
const finalizedDafId = ref('')
const existingDraftId = ref('')
const finalizeConfirmOpen = ref(false)
const stockReceiptMed = ref<{ medicationId: string; name: string } | null>(null)
const receiptForm = reactive({ lotNumber: '', expiresOn: '', qty: 1 })
const pendingAfterReceipt = ref<'preview' | 'finalize' | ''>('')
const contextClientName = ref('')
const contextPetName = ref('')
const petFoodChain = ref('')
const petDomicile = ref('')
const petSpecies = ref('')

const clientUserId = computed(() => String(route.query.clientUserId || ''))
const petId = computed(() => String(route.query.petId || ''))
const visitId = computed(() => String(route.query.visitId || ''))
const dafIdQuery = computed(() => String(route.query.dafId || ''))

const petFoodChainLabel = computed(() => {
  switch (petFoodChain.value) {
    case 'food_producing':
      return t('clients.pet.foodChainFoodProducing')
    case 'excluded_from_food_chain':
      return t('clients.pet.foodChainExcluded')
    case 'companion':
      return t('clients.pet.foodChainCompanion')
    default:
      return ''
  }
})

const petRegulatoryLabel = computed(() => Boolean(petFoodChainLabel.value || petDomicile.value))

const contextLabel = computed(() => {
  if (!clientUserId.value && !petId.value && !visitId.value) return ''
  const parts: string[] = []
  if (contextClientName.value) parts.push(contextClientName.value)
  else if (clientUserId.value) parts.push(t('pharmacy.daf.contextClient'))
  if (contextPetName.value) parts.push(contextPetName.value)
  else if (petId.value) parts.push(t('pharmacy.daf.contextPet'))
  if (visitId.value) parts.push(t('pharmacy.daf.contextVisit'))
  return t('pharmacy.daf.contextBanner', { context: parts.join(' · ') })
})

type Line = {
  med: ProComboboxItem | null
  ammNumber: string
  qty: number
  species: string
  indication: string
  durationDays: number
}

const lines = ref<Line[]>([emptyLine()])

function emptyLine(): Line {
  return { med: null, ammNumber: '', qty: 1, species: '', indication: '', durationDays: 5 }
}

function addLine() {
  lines.value.push(emptyLine())
}

const canSubmit = computed(() =>
  lines.value.every((l) => {
    if (!l.med?.id || !l.ammNumber || !(l.qty > 0)) return false
    if (l.med.raw?.isAntibiotic) {
      return !!l.species.trim() && !!l.indication.trim() && l.durationDays > 0
    }
    return true
  }),
)

function unwrap(res: any) {
  return res?.data ?? res
}

async function loadContext() {
  if (clientUserId.value) {
    try {
      const res = await $fetch<any>(`/api/clients/${clientUserId.value}`)
      const c = unwrap(res)
      contextClientName.value = c?.fullName || c?.email || ''
    }
    catch {
      contextClientName.value = ''
    }
  }
  if (petId.value) {
    try {
      const res = await $fetch<any>(`/api/pets/${petId.value}`)
      const pet = unwrap(res)
      if (!contextPetName.value) contextPetName.value = pet?.name || ''
      petFoodChain.value = pet?.foodChainStatus || ''
      petDomicile.value = pet?.domicileLocation || ''
      petSpecies.value = pet?.species || ''
    }
    catch {
      /* keep list-based name if any */
    }
  }
}

function applyDraftDoc(doc: any) {
  if (!doc?.id || doc.status !== 'draft' || !Array.isArray(doc.items) || !doc.items.length) return false
  existingDraftId.value = String(doc.id)
  if (!clientUserId.value && doc.clientUserId) {
    /* query takes precedence; names filled via loadContext when ids present */
  }
  lines.value = doc.items.map((it: any) => ({
    med: {
      id: String(it.medicationId),
      label: String(it.medicationName || it.medicationId),
      hint: it.medicationCnk || undefined,
      badge: it.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
      raw: {
        id: it.medicationId,
        isAntibiotic: !!it.isAntibiotic,
        cnk: it.medicationCnk,
        name: it.medicationName,
      },
    },
    ammNumber: String(it.ammNumber || ''),
    qty: Number(it.qty) || 1,
    species: '',
    indication: '',
    durationDays: 5,
  }))
  for (let i = 0; i < lines.value.length; i++) {
    const raw = doc.items[i]?.vamregPayload
    if (!raw || typeof raw !== 'object') continue
    lines.value[i].species = String(raw.species || '')
    lines.value[i].indication = String(raw.indication || '')
    lines.value[i].durationDays = Number(raw.durationDays) || 5
  }
  return true
}

async function loadDraftFromVisit() {
  if (!visitId.value) return
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/for-visit', { query: { visitId: visitId.value } })
    applyDraftDoc(unwrap(res))
  }
  catch (e: any) {
    const status = e?.statusCode || e?.status || e?.response?.status
    if (status !== 404) {
      error.value = pharmacyErr(e, 'pharmacy.daf.error')
    }
  }
}

async function loadDraftById() {
  if (!dafIdQuery.value) return
  try {
    const res = await $fetch<any>(`/api/vet/pharmacy/daf/${dafIdQuery.value}`)
    applyDraftDoc(unwrap(res))
  }
  catch (e: any) {
    const status = e?.statusCode || e?.status || e?.response?.status
    if (status !== 404) {
      error.value = pharmacyErr(e, 'pharmacy.daf.error')
    }
  }
}

onMounted(() => {
  void loadContext()
  void (async () => {
    if (visitId.value) await loadDraftFromVisit()
    else if (dafIdQuery.value) await loadDraftById()
  })()
})

function onMedSelected(line: Line, med: ProComboboxItem | null) {
  if (!med?.raw) return
  const raw = med.raw as Record<string, unknown>
  if (raw.ammNumber && !line.ammNumber.trim()) {
    line.ammNumber = String(raw.ammNumber)
  }
  if (raw.isAntibiotic && !line.species.trim() && petSpecies.value) {
    line.species = petSpecies.value
  }
  if (raw.isAntibiotic && !line.indication.trim()) {
    line.indication = t('clients.consultation.vamregIndicationDefault')
  }
}

watch(
  lines,
  (rows, oldRows) => {
    rows.forEach((line, idx) => {
      const prev = oldRows?.[idx]?.med
      if (line.med && line.med !== prev) {
        onMedSelected(line, line.med)
      }
    })
  },
  { deep: true },
)

async function searchMedications(q: string): Promise<ProComboboxItem[]> {
  const res = await $fetch<any>('/api/vet/pharmacy/medications/search', { query: { q, limit: '20' } })
  const items = unwrap(res)?.items ?? []
  return items.map((m: any) => ({
    id: m.id,
    label: m.name,
    hint: m.cnk,
    badge: m.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
    raw: m,
  }))
}

function buildItems() {
  return lines.value.map((l) => {
    const item: any = {
      medicationId: l.med!.id,
      qty: l.qty,
      ammNumber: l.ammNumber,
      unit: 'unit',
    }
    if (l.med?.raw?.isAntibiotic) {
      item.vamregPayload = {
        species: l.species,
        indication: l.indication,
        durationDays: l.durationDays,
      }
    }
    return item
  })
}

function draftBody() {
  return {
    clientUserId: clientUserId.value || undefined,
    petId: petId.value || undefined,
    visitId: visitId.value || undefined,
    items: buildItems(),
  }
}

function isStockError(e: any): boolean {
  const code = String(
    e?.data?.error?.code || e?.data?.code || e?.response?._data?.error?.code || '',
  ).trim()
  return code === 'stock_insufficient' || code === 'stock_unavailable_valid_lots'
}

function openStockReceipt(medicationId?: string) {
  const want = String(medicationId || '').trim()
  const failing = (want
    ? lines.value.find(l => l.med?.id === want)
    : undefined)
    || lines.value.find(l => l.med?.id)
  if (!failing?.med) return
  stockReceiptMed.value = { medicationId: failing.med.id, name: failing.med.label }
  receiptForm.lotNumber = ''
  receiptForm.expiresOn = ''
  receiptForm.qty = failing.qty > 0 ? failing.qty : 1
}

function stockErrorMedicationId(e: any): string {
  const details = e?.data?.error?.details || e?.data?.details
  return String(details?.medicationId || '').trim()
}

async function submitReceipt() {
  if (!stockReceiptMed.value) return
  busy.value = true
  error.value = ''
  try {
    await $fetch('/api/vet/pharmacy/batches', {
      method: 'POST',
      body: {
        medicationId: stockReceiptMed.value.medicationId,
        lotNumber: receiptForm.lotNumber.trim(),
        expiresOn: receiptForm.expiresOn,
        qty: receiptForm.qty,
        unit: 'unit',
      },
    })
    stockReceiptMed.value = null
    const action = pendingAfterReceipt.value
    pendingAfterReceipt.value = ''
    if (action === 'finalize') {
      await finalize()
    }
    else {
      await runPreview()
    }
  }
  catch (e: any) {
    error.value = pharmacyErr(e, 'pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

async function runPreview() {
  busy.value = true
  error.value = ''
  stockReceiptMed.value = null
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/preview-fefo', { method: 'POST', body: { items: buildItems() } })
    preview.value = unwrap(res)?.lines ?? []
  }
  catch (e: any) {
    if (isStockError(e)) {
      error.value = t('clients.consultation.treatmentsStockError')
      pendingAfterReceipt.value = 'preview'
      openStockReceipt(stockErrorMedicationId(e))
    }
    else {
      error.value = pharmacyErr(e, 'pharmacy.daf.error')
    }
  }
  finally {
    busy.value = false
  }
}

async function upsertDraft() {
  if (visitId.value) {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/for-visit', { method: 'PUT', body: draftBody() })
    return unwrap(res)
  }
  if (existingDraftId.value || dafIdQuery.value) {
    const id = existingDraftId.value || dafIdQuery.value
    const res = await $fetch<any>(`/api/vet/pharmacy/daf/${id}`, {
      method: 'PATCH',
      body: draftBody(),
    })
    return unwrap(res)
  }
  const res = await $fetch<any>('/api/vet/pharmacy/daf', { method: 'POST', body: draftBody() })
  return unwrap(res)
}

async function saveDraft() {
  busy.value = true
  error.value = ''
  try {
    const doc = await upsertDraft()
    await navigateTo(`/daf/${doc.id}`)
  }
  catch (e: any) {
    error.value = pharmacyErr(e, 'pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

function requestFinalize() {
  if (!canSubmit.value) return
  finalizeConfirmOpen.value = true
}

function onFinalizeConfirmOpen(v: boolean) {
  if (!busy.value) finalizeConfirmOpen.value = v
}

async function confirmFinalize() {
  finalizeConfirmOpen.value = false
  await finalize()
}

async function finalize() {
  busy.value = true
  error.value = ''
  stockReceiptMed.value = null
  try {
    const doc = await upsertDraft()
    const fin = await $fetch<any>(`/api/vet/pharmacy/daf/${doc.id}/finalize`, { method: 'POST', body: {} })
    const out = unwrap(fin)
    const id = out?.id || out?.daf?.id || doc.id
    finalizedDafId.value = id
  }
  catch (e: any) {
    if (isStockError(e)) {
      error.value = t('clients.consultation.treatmentsStockError')
      pendingAfterReceipt.value = 'finalize'
      openStockReceipt(stockErrorMedicationId(e))
    }
    else {
      error.value = pharmacyErr(e, 'pharmacy.daf.error')
    }
  }
  finally {
    busy.value = false
  }
}

async function goInvoice() {
  const q = new URLSearchParams()
  if (clientUserId.value) q.set('clientUserId', clientUserId.value)
  if (visitId.value) q.set('visitId', visitId.value)
  if (finalizedDafId.value) q.set('dafId', finalizedDafId.value)
  q.set('mode', 'fromDaf')
  await navigateTo(`/invoicing?${q.toString()}`)
}
</script>

<style scoped>
.daf-line {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--pf-vet-border);
}
.daf-pet-reg {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
  gap: 0.75rem;
  margin: 0.75rem 0 0;
}
.daf-pet-reg dt {
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
  margin: 0 0 0.2rem;
}
.daf-pet-reg dd {
  margin: 0;
  font-weight: 600;
}
.daf-actions { display: flex; gap: 0.75rem; flex-wrap: wrap; }
.daf-receipt-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.75rem;
  margin: 0.75rem 0;
}
.daf-from-consult {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  margin: 0 0 0.5rem;
}
.pro-link {
  color: var(--pf-vet-accent);
  text-decoration: underline;
}
</style>
