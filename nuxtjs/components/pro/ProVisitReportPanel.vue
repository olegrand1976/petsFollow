<template>
  <div class="pro-visit-report" data-testid="visit-report-panel">
    <div class="pro-visit-report__head">
      <div>
        <label class="pro-label">{{ $t('calendar.reportTitle') }}</label>
        <p v-if="visitDateLabel" class="pro-hint" data-testid="visit-report-date">
          {{ $t('calendar.reportVisitDate') }} : {{ visitDateLabel }}
        </p>
        <p class="pro-hint" data-testid="visit-report-flow-hint">
          {{ $t('calendar.reportFlowHint') }}
        </p>
      </div>
      <ProBadge
        v-if="reportStatus === 'final'"
        variant="success"
        data-testid="visit-report-status-final"
      >
        {{ $t('calendar.reportStatusFinal') }}
      </ProBadge>
    </div>

    <details class="visit-report-howto" data-testid="visit-report-howto">
      <summary>{{ $t('calendar.reportHowItWorksTitle') }}</summary>
      <ol>
        <li>{{ $t('calendar.reportHowItWorksStep1') }}</li>
        <li>{{ $t('calendar.reportHowItWorksStep2') }}</li>
        <li>{{ $t('calendar.reportHowItWorksStep3') }}</li>
      </ol>
    </details>

    <div
      v-if="reportAuthors.length > 1"
      class="pro-flex-gap visit-report-authors"
      data-testid="visit-report-authors"
    >
      <ProButton
        v-for="author in reportAuthors"
        :key="author.authorUserId || author.id"
        :variant="selectedReportAuthorId === author.authorUserId ? 'primary' : 'secondary'"
        :test-id="author.mine ? 'visit-report-author-mine' : `visit-report-author-${author.authorUserId}`"
        @click="selectReportAuthor(author)"
      >
        {{ author.mine
          ? $t('calendar.reportAuthorMine')
          : (author.authorFullName || $t('calendar.reportAuthorPeer')) }}
        <span v-if="author.status" class="pro-hint"> · {{ author.status }}</span>
      </ProButton>
    </div>
    <p
      v-if="viewingPeerReport"
      class="pro-hint"
      data-testid="visit-report-peer-readonly"
    >
      {{ $t('calendar.reportReadOnlyPeer') }}
    </p>

    <div class="visit-report-split">
      <!-- Left: notes / dictation -->
      <section class="visit-report-pane" data-testid="visit-report-pane-left">
        <header class="visit-report-pane__header">
          <div>
            <h3 class="visit-report-pane__title">{{ $t('calendar.reportPaneNotesTitle') }}</h3>
            <p class="pro-hint">{{ $t('calendar.reportPaneNotesSubtitle') }}</p>
          </div>
          <div v-if="!reportLocked && !dictating" class="pro-flex-gap visit-report-pane__toolbar">
            <ProButton
              :disabled="reportBusy || hydrating"
              test-id="visit-report-dictate"
              @click="onDictateClick"
            >
              <ProIcon name="mic" :size="16" />
              {{ $t('calendar.dictateAudio') }}
            </ProButton>
            <ProButton
              variant="secondary"
              :disabled="reportBusy || hydrating"
              test-id="visit-report-audio-btn"
              @click="openAudioPicker"
            >
              <ProIcon name="upload_file" :size="16" />
              {{ $t('calendar.audioToText') }}
            </ProButton>
          </div>
        </header>

        <div
          v-if="dictating"
          class="visit-report-recording"
          data-testid="visit-report-recording-banner"
        >
          <ProIcon name="mic" :size="24" class="visit-report-recording__mic" />
          <div class="visit-report-recording__info">
            <strong>{{ $t('calendar.recordingInProgress') }}</strong>
            <span data-testid="visit-report-recording-clock">{{ dictationClock }}</span>
          </div>
          <ProButton
            variant="secondary"
            test-id="visit-report-dictate-stop"
            @click="stopDictation"
          >
            {{ $t('calendar.recordingStop') }}
          </ProButton>
        </div>

        <label class="visually-hidden" for="visit-report-transcript">{{ $t('calendar.reportPaneNotesTitle') }}</label>
        <textarea
          id="visit-report-transcript"
          v-model="reportTranscript"
          class="pro-input visit-report-pane__textarea"
          rows="12"
          data-testid="visit-report-transcript"
          :placeholder="$t('calendar.reportTranscriptHint')"
          :disabled="reportBusy || reportLocked || dictating || hydrating"
          :readonly="reportLocked"
        />
      </section>

      <!-- Right: CR / AI -->
      <section class="visit-report-pane" data-testid="visit-report-pane-right">
        <header class="visit-report-pane__header">
          <div>
            <h3 class="visit-report-pane__title">{{ $t('calendar.reportPaneCrTitle') }}</h3>
            <p class="pro-hint">{{ $t('calendar.reportPaneCrSubtitle') }}</p>
          </div>
          <div v-if="!reportLocked && !dictating" class="pro-flex-gap visit-report-pane__toolbar">
            <ProButton
              :disabled="reportBusy || hydrating || !canImprove"
              :loading="reportBusy && improveInFlight"
              test-id="visit-report-improve"
              @click="improveVisitReport"
            >
              <ProIcon name="auto_awesome" :size="16" />
              {{ $t('calendar.improveReport') }}
            </ProButton>
          </div>
        </header>
        <p
          v-if="!readonly && !viewingPeerReport"
          class="pro-hint"
          data-testid="visit-report-ai-banner"
        >
          {{ $t('calendar.reportAiProposalBanner') }}
        </p>
        <ProMarkdownReportEditor
          v-model="reportBody"
          input-id="visit-report-body"
          textarea-test-id="visit-report-body"
          editor-test-id="visit-report-editor"
          test-id-prefix="visit-report"
          :placeholder="$t('calendar.reportHint')"
          :disabled="reportBusy || reportLocked || dictating || hydrating"
          :readonly="reportLocked"
          :prefer-preview-tick="preferMarkdownPreviewTick"
        />
      </section>
    </div>

    <input
      ref="audioFileInput"
      type="file"
      accept="audio/*,.mp3,.m4a,.wav,.ogg,.webm"
      class="pro-visit-report__file"
      data-testid="visit-report-audio"
      :disabled="reportBusy"
      @change="onReportAudioSelected"
    >

    <details
      class="visit-report-history"
      data-testid="visit-report-history"
    >
      <summary>{{ $t('calendar.reportVersionsTitle') }}</summary>
      <div class="visit-report-history__block" data-testid="visit-report-v0">
        <div class="pro-flex-gap visit-report-history__head">
          <h4>{{ $t('calendar.reportVersionTranscript') }}</h4>
          <ProButton
            v-if="reportPersistedTranscript && !reportLocked && !dictating"
            variant="ghost"
            test-id="visit-report-restore-v0"
            @click="restoreTranscriptVersion"
          >
            {{ $t('calendar.reportRestore') }}
          </ProButton>
        </div>
        <pre v-if="reportPersistedTranscript" class="visit-report-history__text">{{ reportPersistedTranscript }}</pre>
        <p v-else class="pro-hint">{{ $t('calendar.reportVersionEmpty') }}</p>
      </div>
      <div class="visit-report-history__block" data-testid="visit-report-v1">
        <div class="pro-flex-gap visit-report-history__head">
          <h4>{{ $t('calendar.reportVersionImproved') }}</h4>
          <ProButton
            v-if="reportImproved && !reportLocked && !dictating"
            variant="ghost"
            test-id="visit-report-restore-v1"
            @click="restoreImprovedVersion"
          >
            {{ $t('calendar.reportRestore') }}
          </ProButton>
        </div>
        <pre v-if="reportImproved" class="visit-report-history__text">{{ reportImproved }}</pre>
        <p v-else class="pro-hint">{{ $t('calendar.reportVersionEmpty') }}</p>
      </div>
      <div class="visit-report-history__block" data-testid="visit-report-v2">
        <h4>{{ $t('calendar.reportVersionSaved') }}</h4>
        <pre v-if="historySavedBody" class="visit-report-history__text">{{ historySavedBody }}</pre>
        <p v-else class="pro-hint">{{ $t('calendar.reportVersionEmpty') }}</p>
      </div>
    </details>

    <p v-if="dirty && !reportLocked" class="pro-hint" data-testid="visit-report-dirty-hint">
      {{ $t('calendar.reportDirtyHint') }}
    </p>
    <p v-if="reportMsg" class="pro-hint" data-testid="visit-report-msg" role="alert">{{ reportMsg }}</p>

    <div
      v-if="!reportLocked"
      class="pro-visit-report__footer"
      data-testid="visit-report-footer"
    >
      <ProButton
        variant="ghost"
        :disabled="reportBusy || hydrating || dictating || !dirty"
        test-id="visit-report-cancel"
        @click="cancelEdits"
      >
        {{ $t('calendar.reportDiscardEdits') }}
      </ProButton>
      <ProButton
        variant="secondary"
        :disabled="reportBusy || hydrating || dictating || reportStatus === 'final' || !reportBody.trim()"
        test-id="visit-report-finalize"
        @click="finalizeVisitReport"
      >
        {{ $t('calendar.finalizeReport') }}
      </ProButton>
      <ProButton
        :disabled="reportBusy || hydrating || dictating || reportStatus === 'final' || !dirty"
        :loading="reportBusy && saveInFlight"
        test-id="visit-report-save"
        @click="saveVisitReport"
      >
        {{ $t('calendar.saveReport') }}
      </ProButton>
    </div>

    <ProModal v-model:open="audioConsentOpen" :title="$t('calendar.audioConsentTitle')">
      <p class="pro-hint">{{ $t('calendar.audioConsent') }}</p>
      <label class="pro-checkbox-label" data-testid="audio-consent-checkbox-label">
        <input
          v-model="audioConsentChecked"
          type="checkbox"
          class="pro-checkbox"
          data-testid="audio-consent-checkbox"
        >
        {{ $t('calendar.audioConsentClientCheck') }}
      </label>
      <template #footer>
        <ProButton variant="ghost" test-id="audio-consent-cancel" @click="cancelAudioConsent">
          {{ $t('calendar.cancel') }}
        </ProButton>
        <ProButton
          test-id="audio-consent-accept"
          :disabled="!audioConsentChecked"
          @click="acceptAudioConsent"
        >
          {{ $t('calendar.audioConsentAccept') }}
        </ProButton>
      </template>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
import { mapVisitReportFields, persistedHistoryBody } from '~/utils/visitReport'
import { normalizeReportText } from '~/utils/safeMarkdown'
import { probeAudioDurationSec } from '~/utils/audioDuration'
import { useActiveConsultation } from '~/composables/useActiveConsultation'

export type VisitReportAuthor = {
  id?: string
  authorUserId?: string
  authorFullName?: string
  mine?: boolean
  status?: string
  bodyText?: string
  transcriptText?: string
  improvedText?: string
}

const props = defineProps<{
  visitId: string
  /** When true, hide save/improve/finalize/audio (ACL pets.write_clinical). */
  readonly?: boolean
  /** Visit date shown under the title and used to prefix the report body ("Date du : …"). */
  visitScheduledAt?: string
}>()

const emit = defineEmits<{
  saved: []
  finalized: []
  busy: [value: boolean]
}>()

const { t } = useI18n()
const { mapError } = useApiError()
const { formatDate } = useFormatters()

const reportBody = ref('')
const reportPersistedBody = ref('')
const reportTranscript = ref('')
const reportPersistedTranscript = ref('')
const reportImproved = ref('')
const reportStatus = ref('')
const reportBusy = ref(false)
const saveInFlight = ref(false)
const improveInFlight = ref(false)
/** True while initial/visit switch hydrate runs — locks textarea so fill cannot race GET. */
const hydrating = ref(false)
let hydrateSeq = 0
const reportMsg = ref('')
const reportAuthors = ref<VisitReportAuthor[]>([])
const selectedReportAuthorId = ref('')
const audioConsentOpen = ref(false)
const audioConsentChecked = ref(false)
const pendingAudioFile = ref<File | null>(null)
/** Which action the client-audio-consent modal is gating: file upload or live dictation. */
const pendingAction = ref<'file' | 'dictate' | null>(null)
const audioFileInput = ref<HTMLInputElement | null>(null)
/** Bumps ProMarkdownReportEditor into preview after IA / restore (always fires). */
const preferMarkdownPreviewTick = ref(0)

function showMarkdownPreview() {
  preferMarkdownPreviewTick.value += 1
}

const dictating = ref(false)
const dictationSeconds = ref(0)
const MAX_DICTATION_SECONDS = 10 * 60
let dictationTimer: ReturnType<typeof setInterval> | null = null
let mediaRecorder: MediaRecorder | null = null
let mediaStream: MediaStream | null = null
let recordChunks: Blob[] = []
/** In-flight Stop → transcribe; flush waits on this to avoid overwriting the transcript. */
let stopDictationInFlight: Promise<void> | null = null

watch(reportBusy, (busy) => {
  emit('busy', busy)
}, { immediate: true })

function waitUntilReportIdle(timeoutMs = 90_000): Promise<void> {
  if (!reportBusy.value) return Promise.resolve()
  return new Promise((resolve) => {
    const stop = watch(reportBusy, (busy) => {
      if (!busy) {
        stop()
        clearTimeout(timer)
        resolve()
      }
    })
    const timer = setTimeout(() => {
      stop()
      resolve()
    }, timeoutMs)
  })
}

const viewingPeerReport = computed(() => {
  if (!selectedReportAuthorId.value || reportAuthors.value.length === 0) return false
  const selected = reportAuthors.value.find(a => a.authorUserId === selectedReportAuthorId.value)
  return Boolean(selected && !selected.mine)
})

const reportLocked = computed(() =>
  Boolean(props.readonly) || viewingPeerReport.value || reportStatus.value === 'final',
)

const visitDateLabel = computed(() =>
  props.visitScheduledAt ? formatDate(props.visitScheduledAt) : '',
)

const historySavedBody = computed(() =>
  persistedHistoryBody(
    reportPersistedBody.value,
    reportPersistedTranscript.value,
    reportImproved.value,
  ),
)

const dirty = computed(() =>
  reportBody.value !== reportPersistedBody.value
  || reportTranscript.value !== reportPersistedTranscript.value
  || dictating.value,
)

const canImprove = computed(() =>
  Boolean(reportTranscript.value.trim() || reportBody.value.trim()),
)

function restoreTranscriptVersion() {
  if (reportLocked.value) return
  reportTranscript.value = reportPersistedTranscript.value
}

function restoreImprovedVersion() {
  if (reportLocked.value) return
  reportBody.value = reportImproved.value
  showMarkdownPreview()
}

function cancelEdits() {
  if (reportLocked.value || reportBusy.value || dictating.value) return
  reportBody.value = reportPersistedBody.value
  reportTranscript.value = reportPersistedTranscript.value
  reportMsg.value = ''
}

function openAudioPicker() {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final' || reportBusy.value) return
  audioFileInput.value?.click()
}

const dictationClock = computed(() => {
  const m = Math.floor(dictationSeconds.value / 60).toString().padStart(2, '0')
  const s = (dictationSeconds.value % 60).toString().padStart(2, '0')
  return `${m}:${s}`
})

/** Prepends "Date du : …" once real content exists — idempotent (checked via includes()). */
function applyDatePrefixIfNeeded() {
  if (!visitDateLabel.value) return
  const body = reportBody.value
  if (!body.trim()) return
  const prefix = `${t('calendar.reportDatePrefix')} : ${visitDateLabel.value}`
  if (body.includes(prefix)) return
  reportBody.value = `${prefix}\n\n${body}`
}

function applyReportPayload(data: Record<string, unknown> | null | undefined) {
  const mapped = mapVisitReportFields(data)
  reportBody.value = mapped.bodyText
  reportPersistedBody.value = mapped.bodyText
  reportTranscript.value = mapped.transcriptText
  reportPersistedTranscript.value = mapped.transcriptText
  reportImproved.value = mapped.improvedText
  reportStatus.value = mapped.status
}

function reportHasContent(author: VisitReportAuthor | null | undefined) {
  if (!author) return false
  return Boolean(
    (author.bodyText || '').trim()
    || (author.transcriptText || '').trim()
    || (author.improvedText || '').trim(),
  )
}

function applyPeerReport(author: VisitReportAuthor) {
  const mapped = mapVisitReportFields({
    bodyText: author.bodyText,
    transcriptText: author.transcriptText,
    improvedText: author.improvedText,
    status: author.status,
  })
  reportBody.value = mapped.bodyText
  reportPersistedBody.value = mapped.bodyText
  reportTranscript.value = mapped.transcriptText
  reportPersistedTranscript.value = mapped.transcriptText
  reportImproved.value = mapped.improvedText
  reportStatus.value = mapped.status
}

async function loadVisitReports(visitId: string) {
  try {
    const res: any = await $fetch(`/api/visits/${visitId}/reports`)
    const list = (res.data ?? res) as VisitReportAuthor[]
    reportAuthors.value = Array.isArray(list) ? list : []
  } catch {
    reportAuthors.value = []
  }
}

async function loadVisitReport(visitId: string) {
  if (viewingPeerReport.value) return
  try {
    const res: any = await $fetch(`/api/visits/${visitId}/report`)
    applyReportPayload(res.data ?? res)
  } catch {
    applyReportPayload(null)
  }
}

async function hydrateVisitReports(visitId: string) {
  const seq = ++hydrateSeq
  hydrating.value = true
  reportBody.value = ''
  reportPersistedBody.value = ''
  reportTranscript.value = ''
  reportPersistedTranscript.value = ''
  reportImproved.value = ''
  reportStatus.value = ''
  reportMsg.value = ''
  reportAuthors.value = []
  selectedReportAuthorId.value = ''

  try {
    let minePayload: Record<string, unknown> | null = null
    try {
      const res: any = await $fetch(`/api/visits/${visitId}/report`)
      minePayload = (res.data ?? res) as Record<string, unknown>
    }
    catch {
      minePayload = null
    }

    await loadVisitReports(visitId)

    if (seq !== hydrateSeq || props.visitId !== visitId) {
      return
    }

    const mine = reportAuthors.value.find(a => a.mine)
    const peerWithContent = reportAuthors.value.find(a => !a.mine && reportHasContent(a))
    const mineFromGet = minePayload ? mapVisitReportFields(minePayload) : null
    const mineGetHasContent = Boolean(
      mineFromGet
      && (
        mineFromGet.bodyText.trim()
        || mineFromGet.transcriptText.trim()
        || mineFromGet.improvedText.trim()
      ),
    )

    if (mine && reportHasContent(mine)) {
      selectedReportAuthorId.value = mine.authorUserId || ''
      // Prefer list row if GET /report failed or returned empty while /reports has content.
      if (mineGetHasContent) applyReportPayload(minePayload)
      else applyPeerReport(mine)
      return
    }
    if (peerWithContent) {
      selectedReportAuthorId.value = peerWithContent.authorUserId || ''
      applyPeerReport(peerWithContent)
      return
    }
    if (mine?.authorUserId) {
      selectedReportAuthorId.value = mine.authorUserId
    }
    applyReportPayload(minePayload)
  }
  finally {
    if (seq === hydrateSeq) hydrating.value = false
  }
}

function selectReportAuthor(author: VisitReportAuthor) {
  selectedReportAuthorId.value = author.authorUserId || ''
  if (author.mine) {
    void loadVisitReport(props.visitId)
    return
  }
  applyPeerReport(author)
}

async function saveVisitReport(): Promise<boolean> {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return false
  applyDatePrefixIfNeeded()
  reportBusy.value = true
  saveInFlight.value = true
  reportMsg.value = ''
  try {
    const transcript = normalizeReportText(reportTranscript.value)
    const res: any = await $fetch(`/api/visits/${props.visitId}/report`, {
      method: 'PUT',
      body: {
        bodyText: normalizeReportText(reportBody.value),
        transcriptText: transcript,
      },
    })
    applyReportPayload(res.data ?? res)
    reportMsg.value = t('calendar.reportSaved')
    void loadVisitReports(props.visitId)
    emit('saved')
    return true
  } catch (e: any) {
    reportMsg.value = mapError(e)
    return false
  } finally {
    saveInFlight.value = false
    reportBusy.value = false
  }
}

/**
 * Leave-guard save: finalize in-progress dictation (transcribe) then require body.
 * If Stop→transcribe is already running, wait for it before PUT.
 */
async function forceSave(): Promise<boolean> {
  if (dictating.value || stopDictationInFlight) {
    try {
      await stopDictation()
    }
    catch {
      await discardDictation()
      return false
    }
  }
  await waitUntilReportIdle()
  if (!reportBody.value.trim() && !reportTranscript.value.trim()) return false
  if (!reportBody.value.trim() && reportTranscript.value.trim()) {
    reportBody.value = reportTranscript.value
  }
  return saveVisitReport()
}

/**
 * Desk/visibility flush: finalize in-progress dictation (transcribe) then save.
 * Empty body after that is OK (visit still exists for resume).
 */
async function flushForSuspend(): Promise<boolean> {
  if (dictating.value || stopDictationInFlight) {
    try {
      await stopDictation()
    } catch {
      await discardDictation()
      return false
    }
  }
  await waitUntilReportIdle()
  if (!reportBody.value.trim() && !reportTranscript.value.trim()) return true
  if (!reportBody.value.trim() && reportTranscript.value.trim()) {
    reportBody.value = reportTranscript.value
  }
  return saveVisitReport()
}

function isDirty(): boolean {
  return dirty.value
}

function currentBody(): string {
  return reportBody.value
}

function isDictating(): boolean {
  return dictating.value
}

defineExpose({ forceSave, flushForSuspend, isDirty, currentBody, isDictating })

async function improveVisitReport() {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return
  const sourceRaw = reportTranscript.value.trim() || reportBody.value.trim()
  if (!sourceRaw) return
  applyDatePrefixIfNeeded()
  reportBusy.value = true
  improveInFlight.value = true
  reportMsg.value = ''
  try {
    // Persist both panes as-is (do NOT overwrite body with transcript — avoid data loss if IA fails).
    const putRes: any = await $fetch(`/api/visits/${props.visitId}/report`, {
      method: 'PUT',
      body: {
        bodyText: normalizeReportText(reportBody.value),
        transcriptText: normalizeReportText(reportTranscript.value),
      },
    })
    // Resync persisted* so a later improve failure does not leave a false dirty / stale hydrate.
    applyReportPayload(putRes.data ?? putRes)
    emit('saved')
    // Flat BFF path: nested …/report/improve is registered but not matched by rou3
    // when …/report (GET/PUT) is also a leaf — see 03d-visit-report-ai-bff.
    const res: any = await $fetch(`/api/visits/${props.visitId}/report-improve`, {
      method: 'POST',
      body: { sourceText: normalizeReportText(sourceRaw) },
    })
    applyReportPayload(res.data ?? res)
    showMarkdownPreview()
    reportMsg.value = t('calendar.reportImproved')
    void loadVisitReports(props.visitId)
    emit('saved')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    improveInFlight.value = false
    reportBusy.value = false
  }
}

async function finalizeVisitReport() {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return
  applyDatePrefixIfNeeded()
  reportBusy.value = true
  reportMsg.value = ''
  try {
    await $fetch(`/api/visits/${props.visitId}/report`, {
      method: 'PUT',
      body: {
        bodyText: normalizeReportText(reportBody.value),
        transcriptText: normalizeReportText(reportTranscript.value),
      },
    })
    const res: any = await $fetch(`/api/visits/${props.visitId}/report-finalize`, {
      method: 'POST',
    })
    applyReportPayload(res.data ?? res)
    reportMsg.value = t('calendar.reportFinalized')
    void loadVisitReports(props.visitId)
    emit('finalized')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

async function onReportAudioSelected(ev: Event) {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  pendingAudioFile.value = file
  pendingAction.value = 'file'
  audioConsentChecked.value = false
  audioConsentOpen.value = true
}

function onDictateClick() {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final' || dictating.value) return
  pendingAction.value = 'dictate'
  audioConsentChecked.value = false
  audioConsentOpen.value = true
}

function cancelAudioConsent() {
  pendingAudioFile.value = null
  pendingAction.value = null
  audioConsentChecked.value = false
  audioConsentOpen.value = false
}

async function acceptAudioConsent() {
  if (!audioConsentChecked.value) return
  const action = pendingAction.value
  const file = pendingAudioFile.value
  pendingAudioFile.value = null
  pendingAction.value = null
  audioConsentOpen.value = false
  audioConsentChecked.value = false
  if (reportStatus.value === 'final') return
  if (action === 'dictate') {
    await startDictation()
    return
  }
  if (!file) return
  await transcribeAudio(file, file.name, await probeAudioDurationSec(file))
}

async function transcribeAudio(file: File | Blob, filename: string, durationSec = 0) {
  reportBusy.value = true
  reportMsg.value = ''
  try {
    const form = new FormData()
    form.append('audio', file, filename)
    form.append('clientAudioConsent', 'true')
    if (durationSec > 0) {
      form.append('audioDurationSec', String(durationSec))
    }
    const res: any = await $fetch(`/api/visits/${props.visitId}/report-transcribe`, {
      method: 'POST',
      body: form,
    })
    applyReportPayload(res.data ?? res)
    applyDatePrefixIfNeeded()
    reportMsg.value = t('calendar.reportTranscribed')
    void loadVisitReports(props.visitId)
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

function stopMediaStream() {
  mediaStream?.getTracks().forEach(track => track.stop())
  mediaStream = null
}

function stopDictationTimer() {
  if (dictationTimer) clearInterval(dictationTimer)
  dictationTimer = null
  dictationSeconds.value = 0
}

async function startDictation() {
  recordChunks = []
  try {
    mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true })
  } catch {
    reportMsg.value = t('calendar.micPermissionDenied')
    return
  }
  try {
    const mime = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
      ? 'audio/webm;codecs=opus'
      : (MediaRecorder.isTypeSupported('audio/webm')
          ? 'audio/webm'
          : (MediaRecorder.isTypeSupported('audio/mp4') ? 'audio/mp4' : ''))
    mediaRecorder = mime
      ? new MediaRecorder(mediaStream, { mimeType: mime })
      : new MediaRecorder(mediaStream)
    mediaRecorder.ondataavailable = (e) => {
      if (e.data?.size) recordChunks.push(e.data)
    }
    mediaRecorder.start(1000)
  } catch {
    stopMediaStream()
    mediaRecorder = null
    reportMsg.value = t('calendar.micPermissionDenied')
    return
  }
  dictationSeconds.value = 0
  dictationTimer = setInterval(() => {
    dictationSeconds.value++
    if (dictationSeconds.value >= MAX_DICTATION_SECONDS) {
      void stopDictation()
    }
  }, 1000)
  dictating.value = true
}

async function stopDictation() {
  if (stopDictationInFlight) return stopDictationInFlight
  stopDictationInFlight = (async () => {
    const recorder = mediaRecorder
    if (!recorder || recorder.state === 'inactive') {
      stopDictationTimer()
      dictating.value = false
      stopMediaStream()
      return
    }
    const blob = await new Promise<Blob | null>((resolve) => {
      recorder.onstop = () => {
        resolve(recordChunks.length ? new Blob(recordChunks, { type: recorder.mimeType || 'audio/webm' }) : null)
      }
      try {
        recorder.stop()
      } catch {
        resolve(null)
      }
    })
    // Capture duration before clearing the live clock.
    const recordedSec = dictationSeconds.value
    mediaRecorder = null
    stopDictationTimer()
    dictating.value = false
    stopMediaStream()
    if (!blob) {
      reportBusy.value = false
      return
    }
    const ext = blob.type.includes('mp4') ? 'm4a' : 'webm'
    // Prefer live clock; fall back to blob metadata (often missing on MediaRecorder webm).
    let durationSec = recordedSec
    if (durationSec <= 0) {
      durationSec = await probeAudioDurationSec(blob)
    }
    // transcribeAudio owns reportBusy from here (sets true again + clears in finally).
    await transcribeAudio(blob, `dictation.${ext}`, durationSec)
  })().finally(() => {
    stopDictationInFlight = null
  })
  return stopDictationInFlight
}

async function discardDictation() {
  stopDictationTimer()
  if (mediaRecorder && mediaRecorder.state !== 'inactive') {
    try { mediaRecorder.stop() } catch { /* already stopped */ }
  }
  mediaRecorder = null
  recordChunks = []
  dictating.value = false
  stopMediaStream()
  // If a Stop→transcribe was already in flight, don't cancel the HTTP call —
  // forceSave awaits it separately when dictating is false.
}

onBeforeUnmount(() => {
  // Desk suspend already ran flushForSuspend (stop+transcribe). Don't discard mid-flight.
  const { suspendDiscard } = useActiveConsultation()
  if (suspendDiscard.value) {
    stopMediaStream()
    return
  }
  void discardDictation()
})

watch(
  () => props.visitId,
  async (id, prev) => {
    if (prev && dictating.value) {
      await discardDictation()
    }
    if (id) await hydrateVisitReports(id)
    else {
      hydrateSeq += 1
      hydrating.value = false
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.pro-visit-report__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.visit-report-howto {
  margin-bottom: 0.75rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  padding: 0.5rem 0.85rem;
  background: var(--pf-vet-bg, #f8fafc);
}

.visit-report-howto > summary {
  cursor: pointer;
  font-weight: 600;
  color: var(--pf-vet-primary);
}

.visit-report-howto ol {
  margin: 0.5rem 0 0.25rem;
  padding-left: 1.25rem;
  font-size: 0.85rem;
  line-height: 1.5;
}

.visit-report-split {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.85rem;
  margin-top: 0.5rem;
}

@media (min-width: 900px) {
  .visit-report-split {
    grid-template-columns: 1fr 1fr;
    align-items: stretch;
  }
}

.visit-report-pane {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-surface, #fff);
  box-shadow: var(--pf-vet-shadow-sm, none);
  min-height: 0;
}

.visit-report-pane__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.visit-report-pane__title {
  margin: 0;
  font-size: 1rem;
  color: var(--pf-vet-primary);
}

.visit-report-pane__toolbar {
  flex-wrap: wrap;
}

.visit-report-pane__textarea {
  min-height: 280px;
  width: 100%;
  resize: vertical;
  flex: 1;
}

.visit-report-recording {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0.9rem;
  border-radius: var(--pf-vet-radius, 8px);
  background: color-mix(in srgb, var(--pf-vet-alert) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--pf-vet-alert) 40%, transparent);
}

.visit-report-recording__mic {
  color: var(--pf-vet-alert);
  animation: pf-visit-report-pulse 0.9s ease-in-out infinite;
}

.visit-report-recording__info {
  display: flex;
  flex-direction: column;
  flex: 1;
  color: var(--pf-vet-alert);
}

@keyframes pf-visit-report-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.35; }
}

.visit-report-authors {
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
}

.pro-visit-report__footer {
  position: sticky;
  bottom: 0;
  z-index: 2;
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.85rem;
  padding: 0.75rem 0 0.15rem;
  background: linear-gradient(to top, var(--pf-vet-surface, #fff) 70%, transparent);
}

.pro-visit-report__file {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.visit-report-history {
  margin-top: 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  padding: 0.65rem 0.85rem;
  background: var(--pf-vet-bg, #f8fafc);
}

.visit-report-history > summary {
  cursor: pointer;
  font-weight: 600;
  color: var(--pf-vet-primary);
}

.visit-report-history__head {
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.25rem;
}

.visit-report-history__head h4 {
  margin: 0;
}

.visit-report-history__block {
  margin-top: 0.75rem;
}

.visit-report-history__block h4 {
  margin: 0 0 0.35rem;
  font-size: 0.85rem;
}

.visit-report-history__text {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.85rem;
  max-height: 12rem;
  overflow: auto;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
