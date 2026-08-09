<template>
  <div
    class="consultation-workspace"
    data-testid="consultation-workspace"
    :data-stage="nextStepsOpen ? 'hub' : 'report'"
  >
    <!-- Étape CR -->
    <div
      v-if="!nextStepsOpen"
      class="consultation-report"
      data-testid="consultation-report"
    >
      <p v-if="actionError" class="pro-error" role="alert">{{ actionError }}</p>
      <ProVisitReportPanel
        ref="reportPanelRef"
        fill-height
        :visit-id="visitId"
        :visit-scheduled-at="scheduledAt || undefined"
        :readonly="readonly"
        @saved="onReportSaved"
        @finalized="onReportFinalized"
        @busy="onReportBusy"
      />
    </div>

    <!-- Étape hub post-save -->
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
      <p
        v-if="!showPrescriptionCta && !showDafCta && !showInvoiceCta"
        class="pro-hint"
        data-testid="consultation-next-steps-empty"
      >
        {{ $t('clients.consultation.nextStepsEmpty') }}
      </p>
    </div>

    <div class="consultation-workspace__footer" data-testid="consultation-workspace-footer">
      <ProButton
        variant="ghost"
        test-id="consultation-cancel"
        :disabled="reportBusy || closing || readonly || reportHydrating"
        @click="requestLeave"
      >
        {{ $t('common.cancel') }}
      </ProButton>

      <template v-if="nextStepsOpen && !readonly">
        <ProButton
          variant="secondary"
          test-id="consultation-cta-done"
          :disabled="closing"
          @click="finishDone"
        >
          {{ $t('clients.consultation.ctaDone') }}
        </ProButton>
      </template>

      <template v-else-if="reportSaved && !readonly && !nextStepsOpen">
        <ProButton
          variant="secondary"
          test-id="consultation-goto-next-steps"
          :disabled="reportBusy || closing"
          @click="openNextSteps"
        >
          {{ $t('clients.consultation.nextStepsTitle') }}
        </ProButton>
      </template>
    </div>
  </div>

  <ProModal
    :open="nextPromptOpen"
    size="md"
    :title="$t('clients.consultation.nextPromptTitle')"
    test-id="consultation-next-prompt"
    @update:open="onNextPromptOpen"
  >
    <p class="pro-hint">{{ $t('clients.consultation.nextPromptHint') }}</p>
    <template #footer>
      <ProButton
        variant="ghost"
        test-id="consultation-next-stay"
        @click="nextPromptOpen = false"
      >
        {{ $t('clients.consultation.nextPromptStay') }}
      </ProButton>
      <ProButton
        test-id="consultation-next-continue"
        @click="openNextSteps"
      >
        {{ $t('clients.consultation.nextPromptContinue') }}
      </ProButton>
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
    <p
      class="pro-hint"
      data-testid="consultation-leave-hint"
      :data-leave-mode="leaveDiscardMode"
    >{{ leaveHintText }}</p>
    <p v-if="leaveError" class="pro-error" role="alert">{{ leaveError }}</p>
    <template #footer>
      <div class="consultation-leave-actions" data-testid="consultation-leave-actions">
        <ProButton
          variant="ghost"
          test-id="consultation-leave-stay"
          :disabled="leaveBusy"
          @click="leavePromptOpen = false"
        >
          {{ $t('clients.consultation.leaveStay') }}
        </ProButton>
        <div class="consultation-leave-actions__end">
          <ProButton
            variant="secondary"
            test-id="consultation-leave-discard"
            :disabled="leaveBusy"
            @click="confirmDiscardLeave"
          >
            {{ leaveDiscardText }}
          </ProButton>
          <ProButton
            test-id="consultation-leave-save"
            :loading="leaveBusy"
            @click="confirmSaveLeave"
          >
            {{ $t('clients.consultation.leaveSave') }}
          </ProButton>
        </div>
      </div>
    </template>
  </ProModal>
</template>

<script setup lang="ts">
import { useActiveConsultation } from '~/composables/useActiveConsultation'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

/** Miroir de `VisitReportSavedOrigin` (ProVisitReportPanel). */
type ReportSavedOrigin = 'save' | 'improve' | 'flush'

type ReportPanelExpose = {
  forceSave: () => Promise<boolean>
  flushForSuspend: () => Promise<boolean>
  isDirty: () => boolean
  currentBody: () => string
  isDictating: () => boolean
}

const props = defineProps<{
  visitId: string
  clientId: string
  petId: string
  scheduledAt?: string
  /** RDV agenda / historique : ne pas annuler la visite à la fermeture. */
  preserveVisit?: boolean
  readonly?: boolean
}>()

const emit = defineEmits<{
  closed: []
  stage: [stage: 'report' | 'hub']
  'leave-prompt': [open: boolean]
  busy: [value: boolean]
}>()

const { t } = useI18n()
const { mapError } = useApiError()
const runtimeConfig = useRuntimeConfig()
const active = useActiveConsultation()
const {
  pharmacyEnabled,
  billitEnabled,
  invoicingUiEnabled,
  visitId: flowVisitId,
  petId: flowPetId,
  clientId: flowClientId,
  scheduledAt: flowScheduledAt,
  reportSaved,
  preserveVisit: flowPreserveVisit,
  reportBusy,
  reset,
  afterSaved,
  setReportBusy,
  discardIfUnsaved,
  markDone,
  dafWizardPath,
  invoicingPath,
  prescriptionWizardPath,
} = useConsultationFlow()

const { canPractice } = usePracticePerms()
const canWritePharmacy = computed(() => canPractice('pharmacy.write'))
const canWriteClinical = computed(() => canPractice('pets.write_clinical'))
const prescriptionsEnabled = computed(() => isPublicFlagOn(runtimeConfig.public.prescriptionsEnabled))
const showDafCta = computed(() => !props.readonly && pharmacyEnabled.value && canWritePharmacy.value)
// `clients.write` : même permission que la création de document côté /invoicing —
// sans elle le CTA mène à un formulaire inaccessible (cas secrétaire).
const showInvoiceCta = computed(() => (
  !props.readonly && invoicingUiEnabled && billitEnabled.value && canPractice('clients.write')
))
const showPrescriptionCta = computed(() => !props.readonly && prescriptionsEnabled.value && canWriteClinical.value)

const actionError = ref('')
const closing = ref(false)
const nextStepsOpen = ref(false)
const nextPromptOpen = ref(false)
const leavePromptOpen = ref(false)
const leaveBusy = ref(false)
const leaveError = ref('')
const reportHydrating = ref(false)
const reportPanelRef = ref<ReportPanelExpose | null>(null)
const { user } = useProUser()

const leaveResolve = ref<((ok: boolean) => void) | null>(null)

/** Mode leave : orphelin walk-in vs quitter sans sauver des edits post-save. */
const leaveDiscardMode = computed<'cancel' | 'abandon'>(() => (
  reportSaved.value ? 'abandon' : 'cancel'
))
const leaveHintText = computed(() => (
  leaveDiscardMode.value === 'abandon'
    ? t('clients.consultation.leaveHintDirty')
    : t('clients.consultation.leaveHint')
))
const leaveDiscardText = computed(() => (
  leaveDiscardMode.value === 'abandon'
    ? t('clients.consultation.leaveDiscardDirty')
    : t('clients.consultation.leaveDiscard')
))

/**
 * Recopie les props dans le flow, sans toucher à l'état d'écran. Sur une reprise
 * agenda, la date et l'animal sont résolus par des requêtes qui atterrissent
 * après la première peinture : les refléter ne veut pas dire « autre
 * consultation ».
 */
function syncFlowProps() {
  flowVisitId.value = props.visitId
  flowClientId.value = props.clientId
  flowPetId.value = props.petId
  flowScheduledAt.value = props.scheduledAt || ''
  flowPreserveVisit.value = !!props.preserveVisit
  active.syncActiveVisit({
    visitId: props.visitId,
    clientId: props.clientId,
    petId: props.petId,
    keepVisit: !!props.preserveVisit,
  })
}

/** Nouvelle consultation : repartir d'un écran vierge, dialogues fermés. */
function bindFlow() {
  syncFlowProps()
  reportSaved.value = false
  nextStepsOpen.value = false
  nextPromptOpen.value = false
  leavePromptOpen.value = false
  actionError.value = ''
  closing.value = false
  void hydrateResumeSaved(props.visitId)
}

/** Mark reportSaved when server already has persisted CR (avoid orphan cancel on leave). */
async function hydrateResumeSaved(id: string) {
  reportHydrating.value = true
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
  finally {
    reportHydrating.value = false
  }
}

// L'identité de la consultation seule justifie de tout remettre à zéro.
watch(
  () => [props.visitId, props.clientId] as const,
  () => {
    if (!props.visitId || !props.clientId) return
    bindFlow()
  },
  { immediate: true },
)

// Métadonnées résolues après coup : les refléter sans fermer un dialogue ouvert
// — sinon la confirmation de fermeture disparaît sous les doigts du véto.
watch(
  () => [props.petId, props.scheduledAt, props.preserveVisit] as const,
  () => {
    if (!props.visitId || !props.clientId) return
    syncFlowProps()
  },
)

watch(nextStepsOpen, (v) => {
  emit('stage', v ? 'hub' : 'report')
})

watch(leavePromptOpen, (v) => {
  emit('leave-prompt', v)
})

watch(reportBusy, (v) => {
  emit('busy', v)
})

async function flushReport() {
  const panel = reportPanelRef.value
  if (!props.visitId || !panel) return
  const ok = await panel.flushForSuspend()
  if (!ok) {
    throw new Error('consultation_flush_failed')
  }
}

function onBeforeUnload(e: BeforeUnloadEvent) {
  if (props.readonly || closing.value || reportHydrating.value) return
  if (!needsLeavePrompt()) return
  e.preventDefault()
  e.returnValue = ''
}

onMounted(() => {
  active.registerFlush(flushReport)
  if (import.meta.client) {
    window.addEventListener('beforeunload', onBeforeUnload)
  }
})

onBeforeUnmount(() => {
  if (import.meta.client) {
    window.removeEventListener('beforeunload', onBeforeUnload)
  }
  active.registerFlush(null)
  active.syncActiveVisit(null)
  if (active.suspendDiscard.value) {
    active.suspendDiscard.value = false
    reset()
    return
  }
})

function panelIsDirty(): boolean {
  try {
    return Boolean(reportPanelRef.value?.isDirty?.())
  }
  catch {
    return false
  }
}

/** Quitter nécessite un prompt si CR jamais sauvé, ou éditions non persistées après save. */
function needsLeavePrompt(): boolean {
  return !reportSaved.value || panelIsDirty()
}

onBeforeRouteLeave((_to, _from, next) => {
  if (props.readonly || !props.visitId || closing.value) {
    next()
    return
  }
  if (!needsLeavePrompt()) {
    next()
    return
  }
  void requestLeaveAsync().then((ok) => next(ok))
})

function requestLeaveAsync(): Promise<boolean> {
  if (leaveResolve.value) {
    // Prompt déjà ouvert — chaîner sans écraser le settle existant.
    return new Promise((resolve) => {
      const prev = leaveResolve.value!
      leaveResolve.value = (ok: boolean) => {
        prev(ok)
        resolve(ok)
      }
    })
  }
  return new Promise((resolve) => {
    leaveResolve.value = resolve
    leaveError.value = ''
    leavePromptOpen.value = true
  })
}

function settleLeave(ok: boolean) {
  const resolve = leaveResolve.value
  leaveResolve.value = null
  resolve?.(ok)
}

function requestLeave() {
  if (reportBusy.value || closing.value || props.readonly || reportHydrating.value) return
  if (leavePromptOpen.value) return
  if (needsLeavePrompt()) {
    leaveError.value = ''
    leavePromptOpen.value = true
    return
  }
  void finishClose(false)
}

function onLeavePromptOpen(v: boolean) {
  if (!v && !leaveBusy.value) {
    leavePromptOpen.value = false
    settleLeave(false)
  }
}

async function finishClose(discard: boolean) {
  closing.value = true
  const fromRouteGuard = leaveResolve.value != null
  try {
    if (discard) {
      await discardIfUnsaved()
    }
    const email = user.value?.email
    if (email) active.clearResumeForEmail(email)
    reset()
    active.syncActiveVisit(null)
    settleLeave(true)
    emit('closed')
    // Route leave : laisser next() poursuivre. Sinon le host (modale) gère la fermeture.
    if (fromRouteGuard) {
      /* no-op */
    }
  }
  finally {
    closing.value = false
  }
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

function onReportSaved(origin: ReportSavedOrigin = 'save') {
  afterSaved()
  if (origin !== 'save') return
  nextPromptOpen.value = true
}

function onReportFinalized() {
  afterSaved()
  openNextSteps()
}

function openNextSteps() {
  nextPromptOpen.value = false
  nextStepsOpen.value = true
}

function onNextPromptOpen(v: boolean) {
  if (!v) nextPromptOpen.value = false
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

defineExpose({
  requestLeave,
  isBusy: () => reportBusy.value || closing.value || leaveBusy.value || reportHydrating.value,
})
</script>

<style scoped>
.consultation-workspace {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  gap: 0.65rem;
}

.consultation-report {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  overflow: hidden;
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

.consultation-workspace__footer {
  flex-shrink: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  justify-content: flex-end;
  padding-top: 0.35rem;
  border-top: 1px solid var(--pf-vet-border);
}

.consultation-leave-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  width: 100%;
}

.consultation-leave-actions__end {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-left: auto;
}
</style>
