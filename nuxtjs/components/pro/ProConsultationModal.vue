<template>
  <ProModal
    :open="open"
    :size="visitId ? 'full' : 'md'"
    :title="$t('clients.consultation.title')"
    test-id="consultation-modal"
    :prevent-close="reportBusy || closing || leavePromptOpen"
    @update:open="onOpenUpdate"
  >
    <!-- Étape A : choisir l'animal puis démarrer la visite -->
    <div v-if="!visitId" class="consultation-setup pro-form" data-testid="consultation-setup">
      <p class="pro-hint">{{ $t('clients.consultation.hint') }}</p>
      <p v-if="loadError" class="pro-error" role="alert">{{ loadError }}</p>
      <p v-if="flowError" class="pro-error" role="alert">{{ flowError }}</p>

      <div class="pro-field">
        <label class="pro-label" for="consultation-pet">{{ $t('clients.consultation.pet') }}</label>
        <select
          id="consultation-pet"
          v-model="selectedPetId"
          class="pro-select"
          data-testid="consultation-pet-select"
          :disabled="petsLoading || starting || !!resumeVisitId"
        >
          <option value="" disabled>{{ $t('clients.consultation.petPlaceholder') }}</option>
          <option v-for="p in pets" :key="p.id" :value="p.id">
            {{ p.name }}{{ p.species ? ` (${p.species})` : '' }}
          </option>
        </select>
      </div>

      <ProInput
        v-model="notes"
        :label="$t('clients.consultation.notes')"
        test-id="consultation-notes"
      />
    </div>

    <!-- Étape B : CR médical + traitements DAF -->
    <div v-else class="consultation-report" data-testid="consultation-report">
      <p v-if="actionError" class="pro-error" role="alert">{{ actionError }}</p>
      <ProVisitReportPanel
        ref="reportPanelRef"
        fill-height
        :visit-id="visitId"
        :visit-scheduled-at="scheduledAt"
        @saved="onReportSaved"
        @finalized="onReportSaved"
        @busy="onReportBusy"
      />
      <ProConsultationTreatmentsPanel
        v-if="showDafCta && visitId && reportSaved"
        ref="treatmentsRef"
        :visit-id="visitId"
        :client-user-id="flowClientId || props.clientId"
        :pet-id="petId || selectedPetId"
        :pet-species="selectedPetSpecies"
        show-steps-hint
        @draft-saved="onTreatmentsDraftSaved"
        @finalized="onTreatmentsFinalized"
      />
    </div>

    <template #footer>
      <ProButton
        variant="ghost"
        test-id="consultation-cancel"
        :disabled="reportBusy || closing"
        @click="close"
      >
        {{ $t('common.cancel') }}
      </ProButton>

      <template v-if="!visitId">
        <ProButton
          test-id="consultation-start"
          :loading="starting"
          :disabled="!selectedPetId || starting || petsLoading"
          @click="start"
        >
          {{ $t('clients.consultation.start') }}
        </ProButton>
      </template>

      <template v-else-if="reportSaved">
        <div
          v-if="showDafCta"
          class="consultation-footer-steps"
          data-testid="consultation-daf-steps"
        >
          <span :class="{ 'is-active': dafStep === 'treatments' }">{{ $t('clients.consultation.stepTreatments') }}</span>
          <span aria-hidden="true">·</span>
          <span :class="{ 'is-active': dafStep === 'preview' }">{{ $t('clients.consultation.stepPreview') }}</span>
          <span aria-hidden="true">·</span>
          <span :class="{ 'is-active': dafStep === 'finalize' }">{{ $t('clients.consultation.stepFinalize') }}</span>
          <span aria-hidden="true">·</span>
          <span :class="{ 'is-active': dafStep === 'invoice' }">{{ $t('clients.consultation.stepInvoice') }}</span>
        </div>
        <ProButton
          v-if="showDafCta && treatmentsHasLines && !treatmentsFinalized"
          test-id="consultation-finalize-daf"
          :disabled="closing || treatmentsBusy"
          :loading="treatmentsBusy"
          @click="finalizeTreatments"
        >
          {{ $t('clients.consultation.treatmentsFinalize') }}
        </ProButton>
        <ProButton
          v-if="showDafCta && treatmentsFinalized && showInvoiceCta"
          test-id="consultation-cta-invoice-from-daf"
          :disabled="closing"
          @click="goInvoiceFromDaf"
        >
          {{ $t('clients.consultation.ctaInvoice') }}
        </ProButton>
        <ProButton
          v-if="showDafCta"
          variant="secondary"
          test-id="consultation-cta-daf"
          :disabled="closing"
          @click="goDaf"
        >
          {{ $t('clients.consultation.ctaDafInvoice') }}
        </ProButton>
        <ProButton
          v-if="showInvoiceCta && !treatmentsFinalized"
          test-id="consultation-cta-invoice"
          :disabled="closing"
          @click="goInvoice"
        >
          {{ $t('clients.consultation.ctaInvoice') }}
        </ProButton>
        <ProButton
          variant="secondary"
          test-id="consultation-cta-done"
          :disabled="closing"
          @click="finishDone"
        >
          {{ $t('clients.consultation.ctaDone') }}
        </ProButton>
      </template>
    </template>
  </ProModal>

  <ProModal
    :open="leavePromptOpen"
    size="md"
    :title="$t('clients.consultation.leaveTitle')"
    test-id="consultation-leave-prompt"
    :prevent-close="leaveBusy"
    @update:open="onLeavePromptOpen"
  >
    <p class="pro-hint">{{ leavePromptHint }}</p>
    <p v-if="leaveError" class="pro-error" role="alert">{{ leaveError }}</p>
    <template #footer>
      <ProButton
        variant="ghost"
        test-id="consultation-leave-stay"
        :disabled="leaveBusy"
        @click="leavePromptOpen = false"
      >
        {{ $t('clients.consultation.leaveStay') }}
      </ProButton>
      <ProButton
        variant="secondary"
        test-id="consultation-leave-discard"
        :disabled="leaveBusy"
        @click="confirmDiscardLeave"
      >
        {{ $t('clients.consultation.leaveDiscard') }}
      </ProButton>
      <ProButton
        test-id="consultation-leave-save"
        :loading="leaveBusy"
        @click="confirmSaveLeave"
      >
        {{ leaveSaveLabel }}
      </ProButton>
    </template>
  </ProModal>
</template>

<script setup lang="ts">
import type { ConsultationPet } from '~/composables/useConsultationFlow'
import { useActiveConsultation } from '~/composables/useActiveConsultation'

type ReportPanelExpose = {
  forceSave: () => Promise<boolean>
  flushForSuspend: () => Promise<boolean>
  isDirty: () => boolean
  currentBody: () => string
  isDictating: () => boolean
}

type TreatmentsPanelExpose = {
  isDirty: () => boolean
  hasUnsavedLines: () => boolean
  saveDraft: () => Promise<boolean>
  requestFinalize: () => void
  finalize: () => Promise<void>
  dafId: { value: string }
  finalized: { value: boolean }
  hasLines: { value: boolean }
  preview: { value: unknown[] }
  busy: { value: boolean }
}

const props = defineProps<{
  open: boolean
  clientId: string
  /** Reprise après veille : visite déjà créée. */
  resumeVisitId?: string
  resumePetId?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  closed: []
}>()

const { t } = useI18n()
const { mapError } = useApiError()
const active = useActiveConsultation()
const {
  pharmacyEnabled,
  billitEnabled,
  invoicingUiEnabled,
  visitId,
  petId,
  clientId: flowClientId,
  scheduledAt,
  starting,
  reportSaved,
  reportBusy,
  error: flowError,
  reset,
  startConsultation,
  afterSaved,
  setReportBusy,
  discardIfUnsaved,
  markDone,
  dafWizardPath,
  invoicingPath,
  loadClientPets,
} = useConsultationFlow()

const { canPractice } = usePracticePerms()
const canWritePharmacy = computed(() => canPractice('pharmacy.write'))
const showDafCta = computed(() => pharmacyEnabled.value && canWritePharmacy.value)
const showInvoiceCta = computed(() => invoicingUiEnabled && billitEnabled.value)

const pets = ref<ConsultationPet[]>([])
const petsLoading = ref(false)
const loadError = ref('')
const actionError = ref('')
const selectedPetId = ref('')
const notes = ref('')
const closing = ref(false)
const leavePromptOpen = ref(false)
const leavePromptMode = ref<'report' | 'treatments'>('report')
const leaveBusy = ref(false)
const leaveError = ref('')
const reportPanelRef = ref<ReportPanelExpose | null>(null)
const treatmentsRef = ref<TreatmentsPanelExpose | null>(null)
const treatmentsDafId = ref('')
const treatmentsFinalized = ref(false)
const { user } = useProUser()

const selectedPetSpecies = computed(() => {
  const pid = petId.value || selectedPetId.value
  const p = pets.value.find(row => row.id === pid)
  return p?.species || ''
})

const treatmentsHasLines = computed(() => treatmentsRef.value?.hasLines?.value ?? false)
const treatmentsBusy = computed(() => treatmentsRef.value?.busy?.value ?? false)

const dafStep = computed(() => {
  if (treatmentsFinalized.value) return 'invoice'
  if (treatmentsRef.value?.preview?.value?.length) return 'finalize'
  if (treatmentsHasLines.value) return 'preview'
  return 'treatments'
})

const leavePromptHint = computed(() => {
  if (leavePromptMode.value === 'treatments') {
    return t('clients.consultation.unsavedTreatmentsHint')
  }
  return t('clients.consultation.leaveHint')
})

const leaveSaveLabel = computed(() => {
  if (leavePromptMode.value === 'treatments') {
    return t('clients.consultation.treatmentsSave')
  }
  return t('clients.consultation.leaveSave')
})

async function loadPets() {
  if (!props.clientId) return
  petsLoading.value = true
  loadError.value = ''
  try {
    pets.value = await loadClientPets(props.clientId)
    if (props.resumePetId) {
      selectedPetId.value = props.resumePetId
    }
    else if (pets.value.length === 1) {
      selectedPetId.value = pets.value[0].id
    }
  }
  catch (e: any) {
    loadError.value = mapError(e) || t('clients.consultation.loadPetsError')
    pets.value = []
  }
  finally {
    petsLoading.value = false
  }
}

function applyResume() {
  if (!props.resumeVisitId) return
  visitId.value = props.resumeVisitId
  petId.value = props.resumePetId || ''
  flowClientId.value = props.clientId
  selectedPetId.value = props.resumePetId || ''
  reportSaved.value = false
  active.syncActiveVisit({
    visitId: props.resumeVisitId,
    clientId: props.clientId,
    petId: props.resumePetId || '',
  })
  void hydrateResumeSchedule(props.resumeVisitId, props.resumePetId || '')
  void hydrateResumeSaved(props.resumeVisitId)
}

/** Mark reportSaved when server already has persisted CR (avoid orphan cancel on leave). */
async function hydrateResumeSaved(id: string) {
  try {
    const res: any = await $fetch(`/api/visits/${id}/report`)
    const data = (res?.data ?? res) as Record<string, unknown> | null
    if (!data) return
    const hasContent = Boolean(
      String(data.bodyText || '').trim()
      || String(data.transcriptText || '').trim()
      || String(data.improvedText || '').trim()
      || data.hasAudio === true
      || data.status === 'final',
    )
    if (hasContent) reportSaved.value = true
  }
  catch {
    /* leave-guard still protected by 409 consultation_has_report */
  }
}

/** No GET /visits/:id — resolve date via pet visit list when possible. */
async function hydrateResumeSchedule(id: string, petHint: string) {
  if (!petHint) return
  try {
    const res: any = await $fetch(`/api/pets/${petHint}/visits`)
    const list = Array.isArray(res?.data ?? res) ? (res.data ?? res) : []
    const visit = list.find((v: any) => v?.id === id)
    const at = visit?.scheduledAt || visit?.proposedScheduledAt || visit?.createdAt || ''
    if (at) scheduledAt.value = String(at)
  } catch {
    /* date label optional on resume */
  }
}

// immediate: host monte le modal avec v-if="clientId" déjà open=true —
// sans ça, loadPets ne part jamais au premier open walk-in.
watch(
  () => [props.open, props.clientId, props.resumeVisitId] as const,
  ([isOpen]) => {
    if (!isOpen) return
    reset()
    selectedPetId.value = ''
    notes.value = ''
    actionError.value = ''
    closing.value = false
    leavePromptOpen.value = false
    leavePromptMode.value = 'report'
    leaveError.value = ''
    treatmentsDafId.value = ''
    treatmentsFinalized.value = false
    if (props.resumeVisitId) {
      applyResume()
    }
    void loadPets()
  },
  { immediate: true },
)

/** Prefer pet from openForClient without tearing down an in-progress visit. */
watch(
  () => props.resumePetId,
  (petId) => {
    if (!props.open || props.resumeVisitId || visitId.value) return
    if (petId) selectedPetId.value = petId
  },
)

watch(
  () => [props.open, visitId.value, props.clientId, petId.value || selectedPetId.value] as const,
  ([isOpen, vid, cid, pid]) => {
    if (!isOpen || !vid || !cid) {
      active.syncActiveVisit(null)
      return
    }
    active.syncActiveVisit({ visitId: vid, clientId: cid, petId: String(pid || '') })
  },
)

watch(
  () => treatmentsRef.value?.finalized?.value,
  (v) => {
    if (v) treatmentsFinalized.value = true
  },
)

watch(
  () => treatmentsRef.value?.dafId?.value,
  (v) => {
    if (v) treatmentsDafId.value = v
  },
)

async function flushReport() {
  const panel = reportPanelRef.value
  if (!visitId.value || !panel) return
  const ok = await panel.flushForSuspend()
  if (!ok) {
    throw new Error('consultation_flush_failed')
  }
}

onMounted(() => {
  active.registerFlush(flushReport)
})

onBeforeUnmount(() => {
  active.registerFlush(null)
  active.syncActiveVisit(null)
  if (active.suspendDiscard.value) {
    // Desk lock / visibility — do not cancel the visit.
    active.suspendDiscard.value = false
    reset()
    return
  }
})

onBeforeRouteLeave((_to, _from, next) => {
  if (!props.open || !visitId.value) {
    next()
    return
  }
  if (!reportSaved.value) {
    leavePromptMode.value = 'report'
    leavePromptOpen.value = true
    next(false)
    return
  }
  if (treatmentsRef.value?.isDirty?.()) {
    leavePromptMode.value = 'treatments'
    leavePromptOpen.value = true
    next(false)
    return
  }
  next()
})

async function finishClose(discard: boolean) {
  closing.value = true
  try {
    if (discard) {
      await discardIfUnsaved()
    }
    const email = user.value?.email
    if (email) active.clearResumeForEmail(email)
    reset()
    active.syncActiveVisit(null)
    emit('update:open', false)
    emit('closed')
  }
  finally {
    closing.value = false
  }
}

async function onOpenUpdate(v: boolean) {
  if (!v) {
    if (reportBusy.value || closing.value || leaveBusy.value) return
    if (visitId.value && !reportSaved.value) {
      leavePromptMode.value = 'report'
      leaveError.value = ''
      leavePromptOpen.value = true
      return
    }
    if (visitId.value && reportSaved.value && treatmentsRef.value?.isDirty?.()) {
      leavePromptMode.value = 'treatments'
      leaveError.value = ''
      leavePromptOpen.value = true
      return
    }
    await finishClose(true)
    return
  }
  emit('update:open', v)
}

function onLeavePromptOpen(v: boolean) {
  if (!v && !leaveBusy.value) leavePromptOpen.value = false
}

function close() {
  if (reportBusy.value || closing.value) return
  void onOpenUpdate(false)
}

async function confirmDiscardLeave() {
  leaveBusy.value = true
  leaveError.value = ''
  try {
    leavePromptOpen.value = false
    await finishClose(true)
  }
  finally {
    leaveBusy.value = false
  }
}

async function confirmSaveLeave() {
  leaveBusy.value = true
  leaveError.value = ''
  try {
    if (leavePromptMode.value === 'treatments') {
      const panel = treatmentsRef.value
      if (!panel) {
        leaveError.value = t('clients.consultation.treatmentsEmpty')
        return
      }
      const ok = await panel.saveDraft()
      if (!ok) {
        leaveError.value = t('pharmacy.daf.error')
        return
      }
      leavePromptOpen.value = false
      await finishClose(false)
      return
    }
    const panel = reportPanelRef.value
    if (!panel) {
      leaveError.value = t('clients.consultation.leaveSaveEmpty')
      return
    }
    const ok = await panel.forceSave()
    if (!ok) {
      leaveError.value = t('clients.consultation.leaveSaveEmpty')
      return
    }
    afterSaved()
    try {
      await markDone()
    }
    catch {
      // CR already saved — leave prompt closes even if done fails.
    }
    leavePromptOpen.value = false
    await finishClose(false)
  }
  catch (e: any) {
    leaveError.value = mapError(e) || t('clients.consultation.leaveSaveEmpty')
  }
  finally {
    leaveBusy.value = false
  }
}

async function start() {
  if (!selectedPetId.value || !props.clientId) return
  try {
    const id = await startConsultation({
      clientId: props.clientId,
      petId: selectedPetId.value,
      notes: notes.value,
    })
    if (id) {
      active.syncActiveVisit({
        visitId: id,
        clientId: props.clientId,
        petId: selectedPetId.value,
      })
    }
  }
  catch {
    // error already on flowError
  }
}

function onReportSaved() {
  afterSaved()
}

function onReportBusy(busy: boolean) {
  setReportBusy(busy)
}

function onTreatmentsDraftSaved(id: string) {
  treatmentsDafId.value = id
}

function onTreatmentsFinalized(id: string) {
  treatmentsDafId.value = id
  treatmentsFinalized.value = true
}

async function finalizeTreatments() {
  treatmentsRef.value?.requestFinalize?.()
}

async function goInvoiceFromDaf() {
  const dafId = treatmentsDafId.value || treatmentsRef.value?.dafId?.value || ''
  await closeAndNavigate(invoicingPath({ dafId, mode: 'fromDaf' }))
}

async function closeAndNavigate(path: string) {
  const target = path
  closing.value = true
  try {
    await discardIfUnsaved()
    const email = user.value?.email
    if (email) active.clearResumeForEmail(email)
    reset()
    active.syncActiveVisit(null)
    emit('update:open', false)
    emit('closed')
  }
  finally {
    closing.value = false
  }
  await nextTick()
  await navigateTo(target)
}

async function goDaf() {
  // Always open the wizard so draft lines can be edited / finalized (detail page is read-only).
  await closeAndNavigate(dafWizardPath())
}

async function goInvoice() {
  await closeAndNavigate(invoicingPath({ mode: 'direct' }))
}

async function finishDone() {
  actionError.value = ''
  try {
    await markDone()
    closing.value = true
    try {
      await discardIfUnsaved()
      const email = user.value?.email
      if (email) active.clearResumeForEmail(email)
      reset()
      active.syncActiveVisit(null)
      emit('update:open', false)
      emit('closed')
    }
    finally {
      closing.value = false
    }
  }
  catch (e: any) {
    actionError.value = mapError(e) || t('clients.consultation.markDoneError')
  }
}
</script>

<style scoped>
.consultation-setup {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.consultation-report {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.consultation-footer-steps {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
  margin-right: auto;
  padding-right: 0.75rem;
}
.consultation-footer-steps .is-active {
  color: var(--pf-vet-accent);
  font-weight: 600;
}
</style>
