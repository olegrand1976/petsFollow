<template>
  <div data-testid="prescriptions-detail-page">
    <ProPageHeader
      :title="doc ? `${doc.petName || $t('prescriptions.title')}` : $t('prescriptions.title')"
      :subtitle="statusLabel(doc?.status)"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="prescriptions-detail-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton variant="secondary" test-id="prescriptions-open-pdf" :disabled="loading || !doc || pdfBusy" @click="openPdf">
          {{ pdfBusy ? $t('prescriptions.loading') : $t('prescriptions.openPdf') }}
        </ProButton>
        <ProButton
          v-if="canEditDraft"
          variant="primary"
          test-id="prescriptions-save"
          :disabled="busy"
          @click="save"
        >
          {{ busy ? $t('prescriptions.loading') : $t('prescriptions.save') }}
        </ProButton>
        <ProButton
          v-if="canEditDraft"
          variant="ghost"
          test-id="prescriptions-delete"
          :disabled="busy"
          @click="remove"
        >
          {{ $t('prescriptions.delete') }}
        </ProButton>
        <ProButton
          v-if="canDispenseDaf"
          variant="secondary"
          test-id="prescriptions-dispense-daf"
          :disabled="busy || !hasCatalogLines"
          @click="dispenseDaf"
        >
          {{ $t('prescriptions.dispenseDaf') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <p v-if="loading" class="pro-hint" data-testid="prescriptions-detail-loading">{{ $t('prescriptions.loading') }}</p>
    <p v-if="error" class="pro-alert">{{ error }}</p>
    <p v-if="canDispenseDaf" class="pro-hint" data-testid="prescriptions-dispense-hint">{{ $t('prescriptions.dispenseDafHint') }}</p>

    <ProCard v-if="doc" class="pro-mb-lg">
      <p class="pro-hint">
        {{ doc.practiceName }} · {{ doc.veterinaryName }} · {{ doc.ownerName }}
        · {{ doc.countryCode }} · {{ doc.paperFormat }}
      </p>
      <div v-if="canEditDraft" class="rx-edit">
        <label class="pro-label">{{ $t('prescriptions.linkedVisit') }}</label>
        <select
          v-model="linkedVisitId"
          class="pro-input"
          data-testid="prescriptions-linked-visit"
          :disabled="visitsLoading || busy || aiBusy"
        >
          <option value="">{{ $t('prescriptions.linkedVisitNone') }}</option>
          <option v-for="v in visits" :key="v.id" :value="v.id">
            {{ formatPrescriptionVisitLabel(v, $t('prescriptions.linkedVisitUnavailable')) }}
          </option>
        </select>
        <p v-if="!visitsLoading && doc.petId && !visits.length" class="pro-hint">
          {{ $t('prescriptions.linkedVisitEmpty') }}
        </p>
        <p v-if="linkedVisitOrphan" class="pro-hint" data-testid="prescriptions-visit-orphan">
          {{ $t('prescriptions.linkedVisitUnavailable') }}
        </p>
        <div v-if="linkedVisitId && !linkedVisitOrphan" class="rx-ai">
          <ProButton
            test-id="prescriptions-ai-prefill"
            :disabled="busy"
            :loading="aiBusy"
            @click="aiPrefill"
          >
            <ProIcon name="auto_awesome" :size="16" />
            {{ $t('prescriptions.aiPrefill') }}
          </ProButton>
        </div>
        <label class="pro-label">{{ $t('prescriptions.paperFormat') }}</label>
        <select v-model="paperFormat" class="pro-input" style="max-width:8rem" data-testid="prescriptions-format">
          <option value="A4">A4</option>
          <option value="A5">A5</option>
        </select>
        <label class="pro-label">{{ $t('prescriptions.validUntil') }}</label>
        <input v-model="validUntil" type="date" class="pro-input" data-testid="prescriptions-valid-until" />
        <label class="pro-label">{{ $t('prescriptions.careAdvice') }}</label>
        <textarea v-model="careAdvice" class="pro-input" rows="3" data-testid="prescriptions-care-advice" />
        <label class="pro-label">{{ $t('prescriptions.notes') }}</label>
        <textarea v-model="notes" class="pro-input" rows="2" />
      </div>
      <div v-else>
        <p v-if="doc.visitId" class="pro-hint" data-testid="prescriptions-linked-visit-ro">
          {{ $t('prescriptions.linkedVisit') }}:
          {{ linkedVisitRoLabel }}
        </p>
        <p v-if="doc.validUntil" class="pro-hint">{{ $t('prescriptions.validUntil') }}: {{ doc.validUntil?.slice?.(0, 10) }}</p>
        <p v-if="doc.careAdvice" class="pro-hint" data-testid="prescriptions-care-advice-ro">{{ doc.careAdvice }}</p>
        <p v-if="doc.notes" class="pro-hint">{{ doc.notes }}</p>
      </div>
    </ProCard>

    <ProCard v-if="doc">
      <div v-for="(line, idx) in lines" :key="idx" class="rx-line" :data-testid="`prescriptions-line-${idx}`">
        <template v-if="canEditDraft">
          <div class="rx-line__full">
            <ProCombobox
              :input-id="`rx-detail-med-${idx}`"
              v-model="line.med"
              :placeholder="$t('prescriptions.medSearch')"
              :search-fn="searchMedications"
              @select="(v) => onMedPicked(idx, v)"
            />
          </div>
          <input v-model="line.name" class="pro-input" :placeholder="$t('prescriptions.medName')" />
          <input v-model="line.dosage" class="pro-input" :placeholder="$t('prescriptions.dosage')" />
          <input v-model="line.form" class="pro-input" :placeholder="$t('prescriptions.form')" />
          <input v-model="line.quantity" class="pro-input" :placeholder="$t('prescriptions.quantity')" />
          <input v-model="line.posology" class="pro-input" :placeholder="$t('prescriptions.posology')" />
          <input v-model="line.withdrawal_period" class="pro-input" :placeholder="$t('prescriptions.withdrawal')" />
          <div class="rx-line__full">
            <ProButton variant="ghost" :test-id="`prescriptions-remove-line-${idx}`" @click="removeLine(idx)">
              {{ $t('prescriptions.removeLine') }}
            </ProButton>
          </div>
        </template>
        <template v-else>
          <p><strong>{{ line.name }}</strong> — {{ line.dosage }} · {{ line.form }} · {{ line.quantity }}</p>
          <p class="pro-hint">{{ line.posology }}</p>
        </template>
      </div>
      <ProButton
        v-if="canEditDraft"
        variant="secondary"
        test-id="prescriptions-add-line"
        @click="lines.push(emptyLine())"
      >
        {{ $t('prescriptions.addLine') }}
      </ProButton>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import {
  fetchPetVisitsForPrescription,
  findVisitOption,
  formatPrescriptionVisitLabel,
  isOrphanLinkedVisit,
  type PrescriptionVisitOption,
  unwrapApiData,
  visitIdForSave,
  withEnsuredLinkedVisit,
} from '~/utils/prescription-visit'
import { openPrescriptionPdfBlob } from '~/utils/prescription-pdf'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pets.read' })
const { t } = useI18n()
const { canPractice } = usePracticePerms()
const { mapError } = usePrescriptionError()
const runtimeConfig = useRuntimeConfig()
const route = useRoute()
const busy = ref(false)
const aiBusy = ref(false)
const pdfBusy = ref(false)
const loading = ref(true)
const visitsLoading = ref(false)
const error = ref('')
const doc = ref<any>(null)
const notes = ref('')
const careAdvice = ref('')
const paperFormat = ref('A4')
const validUntil = ref('')
const linkedVisitId = ref('')
const visits = ref<PrescriptionVisitOption[]>([])
const visitsLoadSeq = ref(0)
const lines = ref<any[]>([])
const canEditDraft = computed(() => doc.value?.status === 'draft' && canPractice('pets.write_clinical'))
const pharmacyOn = computed(() => isPublicFlagOn(runtimeConfig.public.pharmacyEnabled))
const canDispenseDaf = computed(() => pharmacyOn.value && canPractice('pharmacy.write') && !!doc.value)
const hasCatalogLines = computed(() => lines.value.some(l => l.ref_medication_id))
const linkedVisitOrphan = computed(() => isOrphanLinkedVisit(visits.value, linkedVisitId.value))
const linkedVisitRoLabel = computed(() => {
  const id = String(doc.value?.visitId || '').trim()
  if (!id) return ''
  const opt = findVisitOption(visits.value, id)
  if (opt) return formatPrescriptionVisitLabel(opt, t('prescriptions.linkedVisitUnavailable'))
  if (visitsLoading.value) return '…'
  return t('prescriptions.linkedVisitUnavailable')
})

function emptyLine() {
  return {
    med: null as ProComboboxItem | null,
    name: '',
    dosage: '',
    form: '',
    quantity: '',
    posology: '',
    withdrawal_period: '',
    cnk: '',
    ref_medication_id: '',
  }
}

function removeLine(idx: number) {
  lines.value.splice(idx, 1)
  if (!lines.value.length) lines.value.push(emptyLine())
}

function onMedPicked(idx: number, v: ProComboboxItem) {
  const line = lines.value[idx]
  if (!line || !v) return
  line.med = v
  const m = (v.raw || {}) as any
  line.name = String(m.name || v.label || '')
  line.cnk = String(m.cnk || '')
  line.ref_medication_id = String(m.id || v.id || '')
  if (m.pharmaceuticalForm && !line.form) line.form = String(m.pharmaceuticalForm)
}

async function searchMedications(q: string): Promise<ProComboboxItem[]> {
  if (!pharmacyOn.value) return []
  const res = await $fetch<any>('/api/vet/pharmacy/medications/search', { query: { q, limit: '20' } })
  const items = unwrapApiData(res)?.items ?? []
  return items.map((m: any) => ({
    id: m.id,
    label: m.name,
    hint: m.cnk,
    badge: m.isAntibiotic ? t('pharmacy.antibioticWarning') : undefined,
    raw: m,
  }))
}

function statusLabel(s?: string) {
  if (s === 'draft') return t('prescriptions.statusDraft')
  if (s === 'signed') return t('prescriptions.statusSigned')
  if (s === 'sent') return t('prescriptions.statusSent')
  if (s === 'archived') return t('prescriptions.statusArchived')
  return s || ''
}

async function loadVisitsForPet(petId: string) {
  if (!petId) {
    visits.value = []
    return
  }
  const seq = ++visitsLoadSeq.value
  visitsLoading.value = true
  try {
    const rows = await fetchPetVisitsForPrescription(petId)
    if (seq !== visitsLoadSeq.value) return
    visits.value = withEnsuredLinkedVisit(rows, linkedVisitId.value)
  }
  catch {
    if (seq !== visitsLoadSeq.value) return
    visits.value = withEnsuredLinkedVisit([], linkedVisitId.value)
  }
  finally {
    if (seq === visitsLoadSeq.value) visitsLoading.value = false
  }
}

function applyDoc(d: any, opts?: { reloadVisits?: boolean }) {
  doc.value = d
  notes.value = d.notes || ''
  careAdvice.value = d.careAdvice || ''
  paperFormat.value = d.paperFormat || 'A4'
  validUntil.value = d.validUntil ? String(d.validUntil).slice(0, 10) : ''
  linkedVisitId.value = d.visitId ? String(d.visitId) : ''
  const meds = Array.isArray(d.medications) ? d.medications : []
  lines.value = meds.length
    ? meds.map((m: any) => ({
        med: m.ref_medication_id
          ? {
              id: String(m.ref_medication_id),
              label: String(m.name || ''),
              hint: m.cnk || undefined,
              raw: { id: m.ref_medication_id, name: m.name, cnk: m.cnk },
            }
          : null,
        name: m.name || '',
        dosage: m.dosage || '',
        form: m.form || '',
        quantity: m.quantity || '',
        posology: m.posology || '',
        withdrawal_period: m.withdrawal_period || '',
        cnk: m.cnk || '',
        ref_medication_id: m.ref_medication_id || '',
      }))
    : [emptyLine()]
  if (opts?.reloadVisits !== false && d.petId) void loadVisitsForPet(String(d.petId))
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await $fetch<any>(`/api/vet/prescriptions/${route.params.id}`)
    applyDoc(unwrapApiData(res))
  }
  catch (e: any) {
    error.value = mapError(e)
  }
  finally {
    loading.value = false
  }
}

function openPdf() {
  if (!doc.value?.id || pdfBusy.value) return
  pdfBusy.value = true
  error.value = ''
  void openPrescriptionPdfBlob(String(doc.value.id))
    .then((mode) => {
      if (mode === 'download') {
        error.value = t('prescriptions.pdfPopupBlocked')
      }
    })
    .catch((e: any) => {
      error.value = mapError(e)
    })
    .finally(() => {
      pdfBusy.value = false
    })
}

function medPayload() {
  return lines.value
    .filter(l => l.name.trim())
    .map((l) => {
      const med: any = {
        name: l.name.trim(),
        dosage: l.dosage.trim(),
        form: l.form.trim(),
        quantity: l.quantity.trim(),
        posology: l.posology.trim(),
        withdrawal_period: l.withdrawal_period.trim(),
      }
      if (l.cnk) med.cnk = l.cnk
      if (l.ref_medication_id) med.ref_medication_id = l.ref_medication_id
      return med
    })
}

function formHasContent() {
  if (careAdvice.value.trim() || notes.value.trim()) return true
  return lines.value.some(l =>
    l.name.trim() || l.dosage.trim() || l.posology.trim() || l.quantity.trim(),
  )
}

function applySuggestion(data: any) {
  const meds = Array.isArray(data?.medications) ? data.medications : []
  lines.value = meds.length
    ? meds.map((m: any) => ({
        med: null,
        name: String(m.name || ''),
        dosage: String(m.dosage || ''),
        form: String(m.form || ''),
        quantity: String(m.quantity || ''),
        posology: String(m.posology || ''),
        withdrawal_period: String(m.withdrawal_period || ''),
        cnk: String(m.cnk || ''),
        ref_medication_id: String(m.ref_medication_id || ''),
      }))
    : [emptyLine()]
  careAdvice.value = typeof data?.careAdvice === 'string' ? data.careAdvice : ''
  notes.value = typeof data?.notes === 'string' ? data.notes : ''
}

async function aiPrefill() {
  if (!linkedVisitId.value || linkedVisitOrphan.value || !canEditDraft.value || aiBusy.value || busy.value) return
  if (formHasContent() && !confirm(t('prescriptions.aiPrefillConfirm'))) return
  const visitToLink = linkedVisitId.value
  aiBusy.value = true
  error.value = ''
  try {
    // Persist visit link without applyDoc (would wipe unsaved form edits).
    if (String(doc.value?.visitId || '') !== visitToLink) {
      await $fetch(`/api/vet/prescriptions/${route.params.id}`, {
        method: 'PATCH',
        body: { visitId: visitToLink },
      })
      if (doc.value) doc.value.visitId = visitToLink
    }
    const res = await $fetch<any>('/api/vet/prescriptions/suggest-from-visit', {
      method: 'POST',
      body: { visitId: visitToLink },
    })
    applySuggestion(unwrapApiData(res))
  }
  catch (e: any) {
    error.value = mapError(e)
  }
  finally {
    aiBusy.value = false
  }
}

async function save() {
  if (busy.value || doc.value?.status !== 'draft') return
  busy.value = true
  error.value = ''
  try {
    const body: any = {
      notes: notes.value,
      careAdvice: careAdvice.value,
      paperFormat: paperFormat.value,
      validUntil: validUntil.value || '',
      medications: medPayload(),
      visitId: visitIdForSave(visits.value, linkedVisitId.value),
    }
    const res = await $fetch<any>(`/api/vet/prescriptions/${route.params.id}`, {
      method: 'PATCH',
      body,
    })
    applyDoc(unwrapApiData(res))
  }
  catch (e: any) {
    error.value = mapError(e)
  }
  finally {
    busy.value = false
  }
}

async function remove() {
  if (busy.value || doc.value?.status !== 'draft') return
  if (!confirm(t('prescriptions.deleteConfirm'))) return
  busy.value = true
  try {
    await $fetch(`/api/vet/prescriptions/${route.params.id}`, { method: 'DELETE' })
    await navigateTo('/prescriptions')
  }
  catch (e: any) {
    error.value = mapError(e)
  }
  finally {
    busy.value = false
  }
}

async function dispenseDaf() {
  if (!canDispenseDaf.value || !hasCatalogLines.value || busy.value) return
  busy.value = true
  error.value = ''
  try {
    if (canEditDraft.value) {
      const body: any = {
        notes: notes.value,
        careAdvice: careAdvice.value,
        paperFormat: paperFormat.value,
        validUntil: validUntil.value || '',
        medications: medPayload(),
        visitId: visitIdForSave(visits.value, linkedVisitId.value),
      }
      const saved = await $fetch<any>(`/api/vet/prescriptions/${route.params.id}`, {
        method: 'PATCH',
        body,
      })
      applyDoc(unwrapApiData(saved))
    }
    const res = await $fetch<any>('/api/vet/pharmacy/daf/from-prescription', {
      method: 'POST',
      body: { prescriptionId: String(route.params.id) },
    })
    const daf = unwrapApiData(res)
    const q = new URLSearchParams()
    if (daf.clientUserId) q.set('clientUserId', String(daf.clientUserId))
    if (daf.petId) q.set('petId', String(daf.petId))
    if (daf.visitId) q.set('visitId', String(daf.visitId))
    if (daf.id) q.set('dafId', String(daf.id))
    await navigateTo(`/daf/nouveau?${q.toString()}`)
  }
  catch (e: any) {
    error.value = mapError(e) || t('pharmacy.daf.error')
  }
  finally {
    busy.value = false
  }
}

onMounted(() => { load() })
</script>

<style scoped>
.rx-line {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.rx-line__full { grid-column: 1 / -1; }
.rx-edit { display: grid; gap: 0.5rem; margin-top: 0.75rem; }
.rx-ai { margin: 0.25rem 0 0.5rem; }
.pro-mb-lg { margin-bottom: 1.25rem; }
.pro-hint { color: var(--pf-vet-muted, #667); }
</style>
