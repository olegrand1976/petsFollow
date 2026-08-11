<template>
  <div class="consult-treatments" data-testid="consultation-treatments">
    <div class="consult-treatments__head">
      <h3 class="consult-treatments__title">{{ $t('clients.consultation.treatmentsTitle') }}</h3>
      <p class="pro-hint">{{ $t('clients.consultation.treatmentsHint') }}</p>
      <p v-if="showStepsHint" class="consult-treatments__steps pro-hint" data-testid="consultation-treatments-steps">
        {{ stepsHintLabel }}
      </p>
    </div>

    <p v-if="error" class="pro-error" role="alert" data-testid="consultation-treatments-error">{{ error }}</p>
    <p v-if="okMsg" class="pro-success" data-testid="consultation-treatments-ok">{{ okMsg }}</p>

    <ProCard
      v-if="protocols.length && !finalized"
      class="consult-treatments__protocols"
      data-testid="consultation-treatments-protocols"
    >
      <h4 class="consult-treatments__protocols-title">{{ $t('clients.consultation.protocolsTitle') }}</h4>
      <div class="consult-treatments__protocols-actions">
        <ProButton
          v-for="proto in protocols"
          :key="proto.id"
          variant="secondary"
          :test-id="`consultation-protocol-${proto.id}`"
          :disabled="busy"
          @click="applyProtocol(proto, false)"
        >
          {{ proto.name }}
        </ProButton>
      </div>
    </ProCard>

    <div
      v-for="(line, idx) in lines"
      :key="idx"
      class="consult-treatments__line"
      :data-testid="`consultation-treatment-line-${idx}`"
    >
      <div>
        <label class="pro-label">{{ $t('pharmacy.stock.medication') }}</label>
        <ProCombobox
          :input-id="`consult-daf-med-${idx}`"
          v-model="line.med"
          :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
          :search-fn="searchMedications"
          :disabled="finalized || busy"
        />
      </div>
      <div>
        <label class="pro-label">{{ $t('pharmacy.daf.amm') }}</label>
        <input
          v-model="line.ammNumber"
          class="pro-input"
          :data-testid="`consultation-treatment-amm-${idx}`"
          :disabled="finalized || busy"
        />
      </div>
      <div>
        <label class="pro-label">{{ $t('pharmacy.stock.qty') }}</label>
        <input
          v-model.number="line.qty"
          type="number"
          min="0.001"
          step="any"
          class="pro-input"
          :data-testid="`consultation-treatment-qty-${idx}`"
          :disabled="finalized || busy"
        />
      </div>
      <template v-if="line.med?.raw?.isAntibiotic">
        <div>
          <label class="pro-label">{{ $t('pharmacy.daf.vamregSpecies') }}</label>
          <input
            v-model="line.species"
            class="pro-input"
            :data-testid="`consultation-treatment-species-${idx}`"
            :disabled="finalized || busy"
          />
        </div>
        <div>
          <label class="pro-label">{{ $t('pharmacy.daf.vamregIndication') }}</label>
          <input
            v-model="line.indication"
            class="pro-input"
            :data-testid="`consultation-treatment-indication-${idx}`"
            :disabled="finalized || busy"
          />
        </div>
        <div>
          <label class="pro-label">{{ $t('pharmacy.daf.vamregDuration') }}</label>
          <input
            v-model.number="line.durationDays"
            type="number"
            min="1"
            class="pro-input"
            :data-testid="`consultation-treatment-duration-${idx}`"
            :disabled="finalized || busy"
          />
        </div>
        <div>
          <label class="pro-label">{{ $t('pharmacy.daf.vamregPosology') }}</label>
          <input
            v-model="line.posology"
            class="pro-input"
            :data-testid="`consultation-treatment-posology-${idx}`"
            :disabled="finalized || busy"
          />
        </div>
      </template>
      <div v-if="lines.length > 1 && !finalized" class="consult-treatments__remove">
        <ProButton variant="ghost" :test-id="`consultation-treatment-remove-${idx}`" :disabled="busy" @click="removeLine(idx)">
          {{ $t('clients.consultation.treatmentsRemoveLine') }}
        </ProButton>
      </div>
    </div>

    <div v-if="!finalized" class="consult-treatments__add">
      <ProButton variant="secondary" test-id="consultation-treatment-add" :disabled="busy" @click="addLine">
        {{ $t('pharmacy.daf.addLine') }}
      </ProButton>
    </div>

    <ProCard
      v-if="stockReceiptMed"
      class="consult-treatments__receipt"
      data-testid="consultation-treatments-receipt"
    >
      <h4>{{ $t('clients.consultation.receiptExpressTitle') }}</h4>
      <p class="pro-hint">{{ stockReceiptMed.name }}</p>
      <div class="consult-treatments__receipt-form">
        <div>
          <label class="pro-label">{{ $t('clients.consultation.receiptLot') }}</label>
          <input v-model="receiptForm.lotNumber" class="pro-input" data-testid="consultation-receipt-lot" :disabled="busy" />
        </div>
        <div>
          <label class="pro-label">{{ $t('clients.consultation.receiptExpiry') }}</label>
          <input v-model="receiptForm.expiresOn" type="date" class="pro-input" data-testid="consultation-receipt-expiry" :disabled="busy" />
        </div>
        <div>
          <label class="pro-label">{{ $t('clients.consultation.receiptQty') }}</label>
          <input v-model.number="receiptForm.qty" type="number" min="0.001" step="any" class="pro-input" data-testid="consultation-receipt-qty" :disabled="busy" />
        </div>
      </div>
      <ProButton
        variant="primary"
        test-id="consultation-receipt-submit"
        :disabled="busy || !receiptForm.lotNumber.trim() || !receiptForm.expiresOn || receiptForm.qty <= 0"
        :loading="busy && busyAction === 'receipt'"
        @click="submitReceipt"
      >
        {{ $t('clients.consultation.receiptSubmit') }}
      </ProButton>
    </ProCard>

    <p v-if="finalized" class="pro-success" data-testid="consultation-treatments-finalized">
      {{ $t('clients.consultation.treatmentsFinalized') }}
    </p>

    <ProCard v-if="preview.length" class="consult-treatments__preview" data-testid="consultation-treatments-fefo">
      <h4>{{ $t('pharmacy.daf.previewTitle') }}</h4>
      <ul>
        <li v-for="(p, i) in preview" :key="i">
          {{ p.medicationName }} · lot {{ p.lotNumber }} · {{ p.qty }} · DLC {{ p.expiresOn }}
        </li>
      </ul>
    </ProCard>

    <div v-if="!finalized" class="consult-treatments__actions">
      <ProButton
        variant="secondary"
        test-id="consultation-treatments-save"
        :disabled="busy || !canSubmit"
        :loading="busy && busyAction === 'save'"
        @click="saveDraft"
      >
        {{ $t('clients.consultation.treatmentsSave') }}
      </ProButton>
      <ProButton
        variant="secondary"
        test-id="consultation-treatments-preview"
        :disabled="busy || !canSubmit"
        :loading="busy && busyAction === 'preview'"
        @click="runPreview"
      >
        {{ $t('clients.consultation.treatmentsPreview') }}
      </ProButton>
      <ProButton
        variant="primary"
        test-id="consultation-treatments-finalize"
        :disabled="busy || !canSubmit"
        :loading="busy && busyAction === 'finalize'"
        @click="requestFinalize"
      >
        {{ $t('clients.consultation.treatmentsFinalize') }}
      </ProButton>
    </div>

    <ProModal
      :open="finalizeConfirmOpen"
      size="md"
      :title="$t('clients.consultation.finalizeConfirmTitle')"
      test-id="consultation-treatments-finalize-confirm"
      :prevent-close="busy"
      @update:open="onFinalizeConfirmOpen"
    >
      <p class="pro-hint">{{ $t('clients.consultation.finalizeConfirmBody') }}</p>
      <template #footer>
        <ProButton variant="ghost" test-id="consultation-finalize-cancel" :disabled="busy" @click="finalizeConfirmOpen = false">
          {{ $t('common.cancel') }}
        </ProButton>
        <ProButton
          variant="primary"
          test-id="consultation-finalize-ok"
          :loading="busy && busyAction === 'finalize'"
          @click="confirmFinalize"
        >
          {{ $t('clients.consultation.finalizeConfirmOk') }}
        </ProButton>
      </template>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import { pharmacyErrorMessage } from '~/utils/pharmacy-error'

type ClinicalProtocol = {
  id: string
  name: string
  description?: string
  lines: Array<{ medicationId: string; qty: number; ammNumber?: string; unit?: string }>
}

const props = defineProps<{
  visitId: string
  clientUserId: string
  petId: string
  petSpecies?: string
  /** Optional one-sitting step hint for parent footer. */
  showStepsHint?: boolean
}>()

const emit = defineEmits<{
  'draft-saved': [dafId: string]
  finalized: [dafId: string]
}>()

const { t } = useI18n()
function pharmacyErr(e: any, fallbackKey: string): string {
  return pharmacyErrorMessage(t, e, fallbackKey)
}

type Line = {
  med: ProComboboxItem | null
  ammNumber: string
  qty: number
  species: string
  indication: string
  durationDays: number
  posology: string
}

function emptyLine(): Line {
  return { med: null, ammNumber: '', qty: 1, species: '', indication: '', durationDays: 5, posology: '' }
}

const lines = ref<Line[]>([emptyLine()])
const busy = ref(false)
const busyAction = ref<'save' | 'preview' | 'finalize' | 'receipt' | ''>('')
const error = ref('')
const okMsg = ref('')
const preview = ref<any[]>([])
const dafId = ref('')
const finalized = ref(false)
const finalizeConfirmOpen = ref(false)
const protocols = ref<ClinicalProtocol[]>([])
const stockReceiptMed = ref<{ medicationId: string; name: string } | null>(null)
const receiptForm = reactive({ lotNumber: '', expiresOn: '', qty: 1 })
const pendingAfterReceipt = ref<'preview' | 'finalize' | ''>('')

let lastSavedSnapshot = ''

const canSubmit = computed(() =>
  lines.value.some(l => l.med?.id) && lines.value.every((l) => {
    if (!l.med?.id) return true
    if (!l.ammNumber || !(l.qty > 0)) return false
    if (l.med.raw?.isAntibiotic) {
      return !!l.species.trim() && !!l.indication.trim() && l.durationDays > 0 && !!l.posology.trim()
    }
    return true
  }),
)

const hasLines = computed(() => lines.value.some(l => l.med?.id))

const stepsHintLabel = computed(() => {
  if (finalized.value) return t('clients.consultation.stepInvoice')
  if (preview.value.length) return t('clients.consultation.stepFinalize')
  if (hasLines.value) return t('clients.consultation.stepPreview')
  return t('clients.consultation.stepTreatments')
})

function unwrap(res: any) {
  return res?.data ?? res
}

function linesSnapshot() {
  return JSON.stringify(lines.value.map(l => ({
    medId: l.med?.id || '',
    ammNumber: l.ammNumber,
    qty: l.qty,
    species: l.species,
    indication: l.indication,
    durationDays: l.durationDays,
    posology: l.posology,
  })))
}

function markClean() {
  lastSavedSnapshot = linesSnapshot()
}

function isDirty(): boolean {
  return linesSnapshot() !== lastSavedSnapshot
}

function hasUnsavedLines(): boolean {
  return hasLines.value && isDirty()
}

function addLine() {
  lines.value.push(emptyLine())
}

function removeLine(idx: number) {
  lines.value.splice(idx, 1)
  if (!lines.value.length) lines.value = [emptyLine()]
}

function onMedSelected(line: Line, med: ProComboboxItem | null) {
  if (!med?.raw) return
  const raw = med.raw as Record<string, unknown>
  if (raw.ammNumber && !line.ammNumber.trim()) {
    line.ammNumber = String(raw.ammNumber)
  }
  if (raw.isAntibiotic) {
    if (!line.species.trim() && props.petSpecies) {
      line.species = props.petSpecies
    }
    if (!line.indication.trim()) {
      line.indication = t('clients.consultation.vamregIndicationDefault')
    }
    if (!line.durationDays || line.durationDays < 1) {
      line.durationDays = 5
    }
    if (!line.posology.trim()) {
      line.posology = t('pharmacy.daf.vamregPosologyDefault')
    }
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
  return lines.value
    .filter(l => l.med?.id)
    .map((l) => {
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
          posology: l.posology,
        }
      }
      return item
    })
}

function applyDoc(doc: any) {
  if (!doc?.id) return
  dafId.value = String(doc.id)
  if (doc.status === 'finalized') {
    finalized.value = true
  }
  if (!Array.isArray(doc.items) || !doc.items.length) {
    markClean()
    return
  }
  lines.value = doc.items.map((it: any) => {
    const payload = typeof it.vamregPayload === 'object' && it.vamregPayload ? it.vamregPayload : {}
    return {
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
          ammNumber: it.ammNumber,
        },
      },
      ammNumber: String(it.ammNumber || ''),
      qty: Number(it.qty) || 1,
      species: String(payload.species || ''),
      indication: String(payload.indication || ''),
      durationDays: Number(payload.durationDays) || 5,
      posology: String(payload.posology || ''),
    } as Line
  })
  markClean()
}

async function loadDraft() {
  if (!props.visitId) return
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/for-visit', { query: { visitId: props.visitId } })
    applyDoc(unwrap(res))
  }
  catch (e: any) {
    const status = e?.statusCode || e?.status || e?.response?.status
    if (status !== 404) {
      error.value = pharmacyErr(e, 'pharmacy.daf.error')
    }
    else {
      markClean()
    }
  }
}

async function loadProtocols() {
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/protocols')
    const items = unwrap(res)?.items ?? unwrap(res) ?? []
    protocols.value = Array.isArray(items) ? items : []
  }
  catch {
    protocols.value = []
  }
}

async function resolveMedItem(medicationId: string, ammNumber?: string): Promise<ProComboboxItem> {
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/medications/search', { query: { q: medicationId, limit: '5' } })
    const items = unwrap(res)?.items ?? []
    const hit = items.find((m: any) => String(m.id) === medicationId)
    if (hit) {
      return {
        id: String(hit.id),
        label: String(hit.name),
        hint: hit.cnk,
        badge: hit.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
        raw: hit,
      }
    }
  }
  catch {
    /* fallback below */
  }
  return {
    id: medicationId,
    label: medicationId,
    raw: { id: medicationId, ammNumber },
  }
}

async function applyProtocol(proto: ClinicalProtocol, replace: boolean) {
  if (!proto.lines?.length) return
  busy.value = true
  error.value = ''
  try {
    const newLines: Line[] = []
    for (const pl of proto.lines) {
      const med = await resolveMedItem(pl.medicationId, pl.ammNumber)
      const line = emptyLine()
      line.med = med
      line.qty = pl.qty || 1
      line.ammNumber = pl.ammNumber || String((med.raw as any)?.ammNumber || '')
      onMedSelected(line, med)
      newLines.push(line)
    }
    if (replace) {
      lines.value = newLines
    }
    else {
      const empties = lines.value.filter(l => !l.med?.id)
      if (empties.length >= newLines.length) {
        let i = 0
        for (const line of lines.value) {
          if (!line.med?.id && i < newLines.length) {
            Object.assign(line, newLines[i++])
          }
        }
      }
      else {
        lines.value.push(...newLines)
      }
    }
    okMsg.value = t('clients.consultation.applyProtocol', { name: proto.name })
  }
  finally {
    busy.value = false
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
  const details = e?.data?.error?.details || e?.data?.details || e?.response?._data?.error?.details
  return String(details?.medicationId || '').trim()
}

async function submitReceipt() {
  if (!stockReceiptMed.value) return
  busy.value = true
  busyAction.value = 'receipt'
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
    busyAction.value = ''
  }
}

async function saveDraft(): Promise<boolean> {
  if (!canSubmit.value) {
    error.value = t('clients.consultation.treatmentsEmpty')
    return false
  }
  busy.value = true
  busyAction.value = 'save'
  error.value = ''
  okMsg.value = ''
  stockReceiptMed.value = null
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/for-visit', {
      method: 'PUT',
      body: {
        visitId: props.visitId,
        clientUserId: props.clientUserId || undefined,
        petId: props.petId || undefined,
        items: buildItems(),
      },
    })
    const doc = unwrap(res)
    applyDoc(doc)
    okMsg.value = t('clients.consultation.treatmentsSaved')
    emit('draft-saved', String(doc.id))
    return true
  }
  catch (e: any) {
    error.value = pharmacyErr(e, 'pharmacy.daf.error')
    return false
  }
  finally {
    busy.value = false
    busyAction.value = ''
  }
}

async function runPreview() {
  if (!canSubmit.value) return
  busy.value = true
  busyAction.value = 'preview'
  error.value = ''
  stockReceiptMed.value = null
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/preview-fefo', {
      method: 'POST',
      body: { items: buildItems() },
    })
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
    preview.value = []
  }
  finally {
    busy.value = false
    busyAction.value = ''
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
  if (!canSubmit.value) return
  busy.value = true
  busyAction.value = 'finalize'
  error.value = ''
  okMsg.value = ''
  stockReceiptMed.value = null
  try {
    const res = await $fetch<any>('/api/vet/pharmacy/daf/for-visit', {
      method: 'PUT',
      body: {
        visitId: props.visitId,
        clientUserId: props.clientUserId || undefined,
        petId: props.petId || undefined,
        items: buildItems(),
      },
    })
    const doc = unwrap(res)
    const fin = await $fetch<any>(`/api/vet/pharmacy/daf/${doc.id}/finalize`, { method: 'POST', body: {} })
    const out = unwrap(fin)
    const id = String(out?.id || doc.id)
    dafId.value = id
    finalized.value = true
    markClean()
    okMsg.value = t('clients.consultation.treatmentsFinalized')
    emit('finalized', id)
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
    busyAction.value = ''
  }
}

watch(() => props.visitId, () => {
  void loadDraft()
}, { immediate: true })

onMounted(() => {
  void loadProtocols()
})

defineExpose({
  dafId: computed(() => dafId.value),
  finalized: computed(() => finalized.value),
  hasLines,
  preview: computed(() => preview.value),
  busy: computed(() => busy.value),
  isDirty,
  hasUnsavedLines,
  saveDraft,
  runPreview,
  requestFinalize,
  finalize,
})
</script>

<style scoped>
.consult-treatments {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--pf-vet-border);
}
.consult-treatments__title {
  margin: 0 0 0.25rem;
  font-size: 1rem;
}
.consult-treatments__steps {
  margin: 0.35rem 0 0;
  font-size: 0.85rem;
}
.consult-treatments__line {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
  gap: 0.75rem;
  margin-bottom: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--pf-vet-border);
}
.consult-treatments__actions,
.consult-treatments__add,
.consult-treatments__protocols-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.75rem;
}
.consult-treatments__protocols-title {
  margin: 0 0 0.35rem;
  font-size: 0.95rem;
}
.consult-treatments__preview,
.consult-treatments__protocols,
.consult-treatments__receipt {
  margin-top: 0.75rem;
}
.consult-treatments__preview h4,
.consult-treatments__receipt h4 {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
}
.consult-treatments__receipt-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}
.pro-success {
  color: var(--pf-vet-accent, #0d9488);
  margin: 0.5rem 0;
}
</style>
