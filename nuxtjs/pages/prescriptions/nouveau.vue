<template>
  <div data-testid="prescriptions-wizard-page">
    <ProPageHeader
      :title="$t('prescriptions.wizardTitle')"
      :subtitle="$t('prescriptions.wizardSubtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="prescriptions-wizard-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <p v-if="error" class="pro-alert" data-testid="prescriptions-wizard-error">{{ error }}</p>

    <ProCard class="pro-mb-lg">
      <label class="pro-label">{{ $t('prescriptions.pet') }}</label>
      <ProCombobox
        input-id="rx-pet"
        v-model="pet"
        :placeholder="$t('prescriptions.petSearch')"
        :search-fn="searchPets"
        data-testid="prescriptions-pet"
      />

      <div v-if="pet?.id" class="rx-visit">
        <label class="pro-label">{{ $t('prescriptions.linkedVisit') }}</label>
        <select
          v-model="linkedVisitId"
          class="pro-input"
          data-testid="prescriptions-linked-visit"
          :disabled="visitsLoading"
        >
          <option value="">{{ $t('prescriptions.linkedVisitNone') }}</option>
          <option v-for="v in visits" :key="v.id" :value="v.id">
            {{ formatPrescriptionVisitLabel(v, $t('prescriptions.linkedVisitUnavailable')) }}
          </option>
        </select>
        <p v-if="!visitsLoading && pet?.id && !visits.length" class="pro-hint" data-testid="prescriptions-no-visits">
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
      </div>

      <div class="rx-format">
        <label class="pro-label">{{ $t('prescriptions.paperFormat') }}</label>
        <select v-model="paperFormat" class="pro-input" data-testid="prescriptions-format">
          <option value="A4">A4</option>
          <option value="A5">A5</option>
        </select>
      </div>
      <label class="pro-label">{{ $t('prescriptions.validUntil') }}</label>
      <input v-model="validUntil" type="date" class="pro-input" data-testid="prescriptions-valid-until" />
      <label class="pro-label">{{ $t('prescriptions.careAdvice') }}</label>
      <textarea v-model="careAdvice" class="pro-input" rows="3" data-testid="prescriptions-care-advice" />
      <label class="pro-label">{{ $t('prescriptions.notes') }}</label>
      <textarea v-model="notes" class="pro-input" rows="2" data-testid="prescriptions-notes" />
    </ProCard>

    <ProCard class="pro-mb-lg">
      <div v-for="(line, idx) in lines" :key="idx" class="rx-line" :data-testid="`prescriptions-line-${idx}`">
        <div class="rx-line__full">
          <label class="pro-label">{{ $t('prescriptions.medSearch') }}</label>
          <ProCombobox
            :input-id="`rx-med-${idx}`"
            v-model="line.med"
            :placeholder="$t('prescriptions.medSearch')"
            :search-fn="searchMedications"
            :data-testid="`prescriptions-med-search-${idx}`"
            @select="(v) => onMedPicked(idx, v)"
          />
        </div>
        <div>
          <label class="pro-label">{{ $t('prescriptions.medName') }}</label>
          <input v-model="line.name" class="pro-input" :data-testid="`prescriptions-med-name-${idx}`" />
        </div>
        <div>
          <label class="pro-label">{{ $t('prescriptions.dosage') }}</label>
          <input v-model="line.dosage" class="pro-input" />
        </div>
        <div>
          <label class="pro-label">{{ $t('prescriptions.form') }}</label>
          <input v-model="line.form" class="pro-input" />
        </div>
        <div>
          <label class="pro-label">{{ $t('prescriptions.quantity') }}</label>
          <input v-model="line.quantity" class="pro-input" />
        </div>
        <div class="rx-line__full">
          <label class="pro-label">{{ $t('prescriptions.posology') }}</label>
          <input v-model="line.posology" class="pro-input" />
        </div>
        <div class="rx-line__full">
          <label class="pro-label">{{ $t('prescriptions.withdrawal') }}</label>
          <input v-model="line.withdrawal_period" class="pro-input" />
        </div>
        <div class="rx-line__full">
          <ProButton variant="ghost" :test-id="`prescriptions-remove-line-${idx}`" @click="removeLine(idx)">
            {{ $t('prescriptions.removeLine') }}
          </ProButton>
        </div>
      </div>
      <ProButton variant="secondary" test-id="prescriptions-add-line" @click="addLine">
        {{ $t('prescriptions.addLine') }}
      </ProButton>
    </ProCard>

    <div class="rx-actions">
      <ProButton variant="primary" test-id="prescriptions-save-draft" :disabled="busy || aiBusy || !canSubmit" @click="saveDraft">
        {{ busy ? $t('prescriptions.loading') : $t('prescriptions.saveDraft') }}
      </ProButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import {
  fetchPetVisitsForPrescription,
  formatPrescriptionVisitLabel,
  isOrphanLinkedVisit,
  type PrescriptionVisitOption,
  unwrapApiData,
  visitIdForSave,
  withEnsuredLinkedVisit,
} from '~/utils/prescription-visit'

definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pets.write_clinical' })
const { t } = useI18n()
const route = useRoute()
const { mapError } = usePrescriptionError()
const runtimeConfig = useRuntimeConfig()
const busy = ref(false)
const aiBusy = ref(false)
const error = ref('')
const pet = ref<ProComboboxItem | null>(null)
const paperFormat = ref('A4')
const validUntil = ref('')
const notes = ref('')
const careAdvice = ref('')
const linkedVisitId = ref(String(route.query.visitId || '').trim())
const visits = ref<PrescriptionVisitOption[]>([])
const visitsLoading = ref(false)
const visitsLoadSeq = ref(0)
const petsCache = ref<any[]>([])
const pharmacyOn = computed(() => isPublicFlagOn(runtimeConfig.public.pharmacyEnabled))
const linkedVisitOrphan = computed(() => isOrphanLinkedVisit(visits.value, linkedVisitId.value))

type Line = {
  med: ProComboboxItem | null
  name: string
  dosage: string
  form: string
  quantity: string
  posology: string
  withdrawal_period: string
  cnk: string
  ref_medication_id: string
}

const lines = ref<Line[]>([emptyLine()])

function emptyLine(): Line {
  return {
    med: null,
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

function addLine() {
  lines.value.push(emptyLine())
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

const canSubmit = computed(() => !!pet.value?.id)

async function ensurePets() {
  if (petsCache.value.length) return
  const res = await $fetch<any>('/api/vet/pets')
  const data = unwrapApiData(res)
  petsCache.value = Array.isArray(data) ? data : (data?.items ?? data?.pets ?? [])
}

async function loadVisitsForPet(petId: string) {
  const seq = ++visitsLoadSeq.value
  visitsLoading.value = true
  try {
    const rows = await fetchPetVisitsForPrescription(petId)
    if (seq !== visitsLoadSeq.value) return
    visits.value = withEnsuredLinkedVisit(rows, linkedVisitId.value)
    // Query visitId for another pet / soft-deleted: keep visible orphan, no AI until cleared.
  }
  catch (e: any) {
    if (seq !== visitsLoadSeq.value) return
    visits.value = withEnsuredLinkedVisit([], linkedVisitId.value)
    error.value = mapError(e)
  }
  finally {
    if (seq === visitsLoadSeq.value) visitsLoading.value = false
  }
}

watch(
  () => pet.value?.id,
  (id, prev) => {
    if (!id) {
      visitsLoadSeq.value += 1
      visits.value = []
      linkedVisitId.value = ''
      return
    }
    const qPet = String(route.query.petId || '').trim()
    const qVisit = String(route.query.visitId || '').trim()
    const keepQueryVisit = qPet === id && qVisit && linkedVisitId.value === qVisit
    // Clear on any pet change (incl. first pick) unless query pet+visit match this animal.
    if (id !== prev && !keepQueryVisit) {
      linkedVisitId.value = ''
    }
    void loadVisitsForPet(id)
  },
)

async function prefillPetFromQuery() {
  const petId = String(route.query.petId || '').trim()
  if (!petId || pet.value?.id) return
  try {
    await ensurePets()
    const match = petsCache.value.find((p: any) => String(p.id) === petId)
    if (!match) return
    pet.value = {
      id: String(match.id),
      label: `${match.name}${match.ownerName ? ` — ${match.ownerName}` : ''}`,
      hint: match.species || undefined,
      raw: match,
    }
  }
  catch {
    /* prefill best-effort */
  }
}

onMounted(() => {
  void prefillPetFromQuery()
})

async function searchPets(q: string): Promise<ProComboboxItem[]> {
  await ensurePets()
  const needle = q.trim().toLowerCase()
  return petsCache.value
    .filter((p) => {
      if (!needle) return true
      const hay = `${p.name || ''} ${p.ownerName || ''} ${p.species || ''}`.toLowerCase()
      return hay.includes(needle)
    })
    .slice(0, 20)
    .map((p) => ({
      id: p.id,
      label: `${p.name}${p.ownerName ? ` — ${p.ownerName}` : ''}`,
      hint: p.species || undefined,
      raw: p,
    }))
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
  if (!linkedVisitId.value || linkedVisitOrphan.value || aiBusy.value || busy.value) return
  if (formHasContent() && !confirm(t('prescriptions.aiPrefillConfirm'))) return
  aiBusy.value = true
  error.value = ''
  try {
    const res = await $fetch<any>('/api/vet/prescriptions/suggest-from-visit', {
      method: 'POST',
      body: { visitId: linkedVisitId.value },
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

async function saveDraft() {
  if (!canSubmit.value || busy.value) return
  busy.value = true
  error.value = ''
  try {
    const body: any = {
      petId: pet.value!.id,
      paperFormat: paperFormat.value,
      notes: notes.value,
      careAdvice: careAdvice.value,
      medications: lines.value
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
        }),
    }
    if (validUntil.value) body.validUntil = validUntil.value
    const vid = visitIdForSave(visits.value, linkedVisitId.value)
    if (vid) body.visitId = vid
    const res = await $fetch<any>('/api/vet/prescriptions', { method: 'POST', body })
    const doc = unwrapApiData(res)
    await navigateTo(`/prescriptions/${doc.id}`)
  }
  catch (e: any) {
    error.value = mapError(e)
  }
  finally {
    busy.value = false
  }
}
</script>

<style scoped>
.rx-line {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--pf-vet-border, #e5e7eb);
}
.rx-line__full { grid-column: 1 / -1; }
.rx-format { margin: 0.75rem 0; max-width: 8rem; }
.rx-visit { margin: 0.75rem 0; }
.rx-ai { margin-top: 0.75rem; }
.rx-actions { display: flex; gap: 0.75rem; }
.pro-mb-lg { margin-bottom: 1.25rem; }
.pro-hint { color: var(--pf-vet-muted, #667); font-size: 0.875rem; }
</style>
