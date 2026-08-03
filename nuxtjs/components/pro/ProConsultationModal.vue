<template>
  <ProModal
    :open="open"
    :size="visitId ? 'full' : 'md'"
    :title="modalTitle"
    test-id="consultation-modal"
    :prevent-close="reportBusy || closing || leavePromptOpen"
    :expandable="!!visitId"
    v-model:expanded="modalExpanded"
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

    <!-- Étape B : CR médical -->
    <div
      v-else-if="!reportSaved"
      class="consultation-report"
      data-testid="consultation-report"
    >
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
    </div>

    <!-- Étape C : hub post-save -->
    <div
      v-else
      class="consultation-next-steps"
      data-testid="consultation-next-steps"
    >
      <p v-if="actionError" class="pro-error" role="alert">{{ actionError }}</p>
      <p class="pro-hint">{{ $t('clients.consultation.nextStepsHint') }}</p>
      <div class="consultation-next-steps__grid">
        <button
          v-if="showPrescriptionCta"
          type="button"
          class="consultation-next-card"
          data-testid="consultation-cta-prescription"
          :disabled="closing"
          @click="goPrescription"
        >
          <ProIcon name="clinical_notes" :size="28" />
          <span class="consultation-next-card__title">
            {{ $t('clients.consultation.ctaPrescription') }}
            <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
          </span>
          <span class="consultation-next-card__desc">{{ $t('clients.consultation.ctaPrescriptionHint') }}</span>
        </button>
        <button
          v-if="showDafCta"
          type="button"
          class="consultation-next-card"
          data-testid="consultation-cta-daf"
          :disabled="closing"
          @click="goDaf"
        >
          <ProIcon name="medication" :size="28" />
          <span class="consultation-next-card__title">{{ $t('clients.consultation.ctaDaf') }}</span>
          <span class="consultation-next-card__desc">{{ $t('clients.consultation.ctaDafHint') }}</span>
        </button>
        <button
          v-if="showInvoiceCta"
          type="button"
          class="consultation-next-card"
          data-testid="consultation-cta-invoice"
          :disabled="closing"
          @click="goInvoice"
        >
          <ProIcon name="receipt_long" :size="28" />
          <span class="consultation-next-card__title">{{ $t('clients.consultation.ctaInvoice') }}</span>
          <span class="consultation-next-card__desc">{{ $t('clients.consultation.ctaInvoiceHint') }}</span>
        </button>
      </div>
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
    <p class="pro-hint">{{ $t('clients.consultation.leaveHint') }}</p>
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
        {{ $t('clients.consultation.leaveSave') }}
      </ProButton>
    </template>
  </ProModal>
</template>

<script setup lang="ts">
import type { ConsultationPet } from '~/composables/useConsultationFlow'
import { useActiveConsultation } from '~/composables/useActiveConsultation'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

type ReportPanelExpose = {
  forceSave: () => Promise<boolean>
  flushForSuspend: () => Promise<boolean>
  isDirty: () => boolean
  currentBody: () => string
  isDictating: () => boolean
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
const runtimeConfig = useRuntimeConfig()
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
  preserveVisit,
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
  prescriptionWizardPath,
  loadClientPets,
} = useConsultationFlow()

const { canPractice } = usePracticePerms()
const canWritePharmacy = computed(() => canPractice('pharmacy.write'))
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))
const prescriptionsEnabled = computed(() => isPublicFlagOn(runtimeConfig.public.prescriptionsEnabled))
const showDafCta = computed(() => pharmacyEnabled.value && canWritePharmacy.value)
const showInvoiceCta = computed(() => invoicingUiEnabled && billitEnabled.value)
const showPrescriptionCta = computed(() => prescriptionsEnabled.value && canWriteClinical.value)

const pets = ref<ConsultationPet[]>([])
const petsLoading = ref(false)
const loadError = ref('')
const actionError = ref('')
const selectedPetId = ref('')
const notes = ref('')
const closing = ref(false)
const leavePromptOpen = ref(false)
const leaveBusy = ref(false)
const leaveError = ref('')
const reportPanelRef = ref<ReportPanelExpose | null>(null)
const modalExpanded = ref(false)
const { user } = useProUser()

const modalTitle = computed(() => {
  if (visitId.value && reportSaved.value) {
    return t('clients.consultation.nextStepsTitle')
  }
  return t('clients.consultation.title')
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
  preserveVisit.value = active.resumeKeepVisit.value
  active.syncActiveVisit({
    visitId: props.resumeVisitId,
    clientId: props.clientId,
    petId: props.resumePetId || '',
    keepVisit: preserveVisit.value,
  })
  void hydrateResumeSchedule(props.resumeVisitId, props.resumePetId || '')
  // RDV agenda : pas de risque d'orphelin (preserveVisit) — on rouvre toujours le CR en édition.
  if (!preserveVisit.value) {
    void hydrateResumeSaved(props.resumeVisitId)
  }
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
    leaveError.value = ''
    modalExpanded.value = false
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
  (petIdHint) => {
    if (!props.open || props.resumeVisitId || visitId.value) return
    if (petIdHint) selectedPetId.value = petIdHint
  },
)

watch(
  () => [props.open, visitId.value, props.clientId, petId.value || selectedPetId.value] as const,
  ([isOpen, vid, cid, pid]) => {
    if (!isOpen || !vid || !cid) {
      active.syncActiveVisit(null)
      return
    }
    active.syncActiveVisit({
      visitId: vid,
      clientId: cid,
      petId: String(pid || ''),
      keepVisit: preserveVisit.value,
    })
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
  await closeAndNavigate(dafWizardPath())
}

async function goInvoice() {
  await closeAndNavigate(invoicingPath({ mode: 'direct' }))
}

async function goPrescription() {
  await closeAndNavigate(prescriptionWizardPath())
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

.consultation-next-steps {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 0.5rem 0 0.25rem;
}

.consultation-next-steps__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
  gap: 0.85rem;
}

.consultation-next-card {
  appearance: none;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.45rem;
  text-align: left;
  padding: 1.1rem 1.15rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-surface, #fff);
  box-shadow: var(--pf-vet-shadow-sm, none);
  color: var(--pf-vet-primary);
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.consultation-next-card:hover:not(:disabled) {
  border-color: var(--pf-vet-accent);
  box-shadow: var(--pf-vet-shadow-hover, var(--pf-vet-shadow-md, none));
}

.consultation-next-card:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.consultation-next-card__title {
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.4rem;
  font-weight: 700;
  font-size: 1rem;
}

.consultation-next-card__desc {
  font-size: 0.85rem;
  color: var(--pf-vet-text-muted, #64748b);
  line-height: 1.35;
}
</style>
