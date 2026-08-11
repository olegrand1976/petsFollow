<template>
  <div
    class="pro-visit-report"
    :class="{ 'pro-visit-report--fill': fillHeight }"
    data-testid="visit-report-panel"
  >
    <!-- Meta + Aide : hors zone scroll (reste visible sous le titre de modale). -->
    <ProVisitReportHead
      :visit-date-label="visitDateLabel"
      :report-status="reportStatus"
    />

    <div class="pro-visit-report__scroll">
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
            <div
              v-if="!reportLocked && !dictating && !transcribeInFlight"
              class="pro-flex-gap visit-report-pane__toolbar"
            >
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
              :loading="transcribeInFlight"
              @click="stopDictation"
            >
              {{ $t('calendar.recordingStop') }}
            </ProButton>
          </div>
          <div
            v-else-if="transcribeInFlight"
            class="visit-report-recording visit-report-recording--busy"
            data-testid="visit-report-transcribing-banner"
            role="status"
            aria-busy="true"
          >
            <ProIcon name="progress_activity" :size="24" class="visit-report-recording__mic" />
            <div class="visit-report-recording__info">
              <strong>{{ $t('calendar.reportTranscribing') }}</strong>
            </div>
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
            <div
              v-if="!reportLocked && !dictating && !transcribeInFlight"
              class="pro-flex-gap visit-report-pane__toolbar"
            >
              <label class="visit-report-lang" data-testid="visit-report-lang">
                <span class="visually-hidden">{{ $t('calendar.reportTargetLang') }}</span>
                <select
                  v-model="targetLocale"
                  class="pro-input visit-report-lang__select"
                  data-testid="visit-report-target-locale"
                  :disabled="reportBusy || hydrating || improveInFlight"
                >
                  <option value="auto">{{ $t('calendar.reportTargetLangAuto') }}</option>
                  <option value="fr">{{ $t('calendar.reportTargetLangFr') }}</option>
                  <option value="nl">{{ $t('calendar.reportTargetLangNl') }}</option>
                  <option value="en">{{ $t('calendar.reportTargetLangEn') }}</option>
                  <option value="es">{{ $t('calendar.reportTargetLangEs') }}</option>
                  <option value="et">{{ $t('calendar.reportTargetLangEt') }}</option>
                  <option value="it">{{ $t('calendar.reportTargetLangIt') }}</option>
                  <!-- uk / ru : masqués tant que les langues ne sont pas annoncées.
                       L'API les accepte déjà (NormalizeVisitReportTargetLocale) et les
                       clés calendar.reportTargetLang{Uk,Ru} existent — il suffit de
                       remettre les deux <option> pour les réactiver. -->
                </select>
              </label>
              <ProButton
                :disabled="reportBusy || hydrating || !canImprove || advancedImproveInFlight"
                :loading="reportBusy && improveInFlight"
                test-id="visit-report-improve"
                @click="improveVisitReport"
              >
                <ProIcon name="auto_awesome" :size="16" />
                {{ $t('calendar.improveReport') }}
              </ProButton>
              <ProButton
                v-if="aiCrAdvancedEnabled"
                variant="secondary"
                :disabled="reportBusy || hydrating || !canImprove || improveInFlight"
                :loading="reportBusy && advancedImproveInFlight"
                test-id="visit-report-improve-advanced"
                @click="improveVisitReportAdvanced"
              >
                <ProIcon name="psychology" :size="16" />
                {{ $t('calendar.improveReportAdvanced') }}
              </ProButton>
            </div>
          </header>
          <div
            v-if="improveInFlight"
            class="visit-report-recording visit-report-recording--busy"
            data-testid="visit-report-improving-banner"
            role="status"
            aria-busy="true"
          >
            <ProIcon name="auto_awesome" :size="24" class="visit-report-recording__mic" />
            <div class="visit-report-recording__info">
              <strong>{{ $t('calendar.reportImproving') }}</strong>
            </div>
          </div>
          <ProAgentLoader
            v-if="advancedImproveInFlight"
            :title="$t('calendar.reportImprovingAdvanced')"
            :waiting-label="$t('calendar.reportImprovingAdvancedWait')"
            :steps="advancedSteps"
            :cancel-label="$t('calendar.reportImproveAdvancedCancel')"
            :status-label="advancedStatusLabel"
            @cancel="cancelAdvancedImprove"
          />
          <p
            v-if="!readonly && !viewingPeerReport"
            class="pro-hint visit-report-pane__ai-hint"
            data-testid="visit-report-ai-banner"
          >
            {{ $t('calendar.reportAiProposalBanner') }}
          </p>
          <ClientOnly>
            <ProRichReportEditor
              v-model="reportBody"
              input-id="visit-report-body"
              textarea-test-id="visit-report-body"
              editor-test-id="visit-report-editor"
              test-id-prefix="visit-report"
              :placeholder="$t('calendar.reportHint')"
              :disabled="reportBusy || reportLocked || dictating || hydrating"
              :readonly="reportLocked"
            />
            <template #fallback>
              <div class="pro-input visit-report-pane__textarea" data-testid="visit-report-editor-fallback">
                {{ reportBody }}
              </div>
            </template>
          </ClientOnly>
          <div
            v-if="reportBody.trim() && !viewingPeerReport"
            class="pro-flex-gap visit-report-export"
            data-testid="visit-report-export"
          >
            <label class="pro-checkbox-label visit-report-export__strip" data-testid="visit-report-strip-citations">
              <input
                v-model="exportStripCitations"
                type="checkbox"
                class="pro-checkbox"
                data-testid="visit-report-strip-citations-check"
              >
              {{ $t('calendar.reportStripCitations') }}
            </label>
            <ProButton
              variant="ghost"
              :disabled="reportBusy || hydrating || exportBusy"
              test-id="visit-report-copy-md"
              @click="copyReportMarkdown"
            >
              <ProIcon name="content_copy" :size="16" />
              {{ $t('calendar.reportCopyMd') }}
            </ProButton>
            <ProButton
              variant="ghost"
              :disabled="reportBusy || hydrating || exportBusy"
              test-id="visit-report-download-md"
              @click="downloadReportMarkdown"
            >
              <ProIcon name="download" :size="16" />
              {{ $t('calendar.reportDownloadMd') }}
            </ProButton>
            <ProButton
              variant="ghost"
              :disabled="reportBusy || hydrating || exportBusy"
              :loading="exportBusy"
              test-id="visit-report-download-pdf"
              @click="downloadReportPdf"
            >
              <ProIcon name="picture_as_pdf" :size="16" />
              {{ $t('calendar.reportDownloadPdf') }}
            </ProButton>
          </div>
          <div
            v-if="aiCrAdvancedEnabled && lastImproveCitations.length && !viewingPeerReport"
            class="visit-report-citations"
            data-testid="visit-report-citations"
          >
            <h4 class="visit-report-citations__title">{{ $t('calendar.reportCitationsTitle') }}</h4>
            <ul class="visit-report-citations__list">
              <li
                v-for="(cite, idx) in lastImproveCitations"
                :key="`cite-${idx}`"
                data-testid="visit-report-citation-item"
              >
                {{ cite }}
              </li>
            </ul>
          </div>
          <div
            v-if="showAiQualityBar"
            class="visit-report-quality"
            data-testid="visit-report-quality"
          >
            <span class="pro-hint">{{ $t('calendar.reportQualityAsk') }}</span>
            <div class="pro-flex-gap">
              <ProButton
                variant="secondary"
                test-id="visit-report-quality-good"
                :disabled="qualityBusy"
                @click="submitAiQuality(true)"
              >
                <ProIcon name="thumb_up" :size="16" />
                {{ $t('calendar.reportQualityGood') }}
              </ProButton>
              <ProButton
                variant="ghost"
                test-id="visit-report-quality-bad"
                :disabled="qualityBusy"
                @click="submitAiQuality(false)"
              >
                <ProIcon name="thumb_down" :size="16" />
                {{ $t('calendar.reportQualityBad') }}
              </ProButton>
            </div>
          </div>
          <label
            v-if="reportStatus === 'final' && !viewingPeerReport && !readonly"
            class="pro-checkbox-label visit-report-reference"
            data-testid="visit-report-reference"
          >
            <input
              v-model="reportIsReference"
              type="checkbox"
              class="pro-checkbox"
              data-testid="visit-report-reference-check"
              :disabled="referenceBusy"
              @change="onReferenceToggle"
            >
            {{ $t('calendar.reportMarkReference') }}
          </label>
        </section>
      </div>

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

    <div
      v-if="!reportLocked"
      class="pro-visit-report__footer"
      data-testid="visit-report-footer"
    >
      <div
        v-if="(dirty && !reportLocked) || reportMsg"
        class="visit-report-status-tags"
        data-testid="visit-report-status-tags"
      >
        <ProBadge
          v-if="dirty && !reportLocked"
          variant="warning"
          data-testid="visit-report-dirty-hint"
        >
          {{ $t('calendar.reportDirtyHint') }}
        </ProBadge>
        <ProBadge
          v-if="reportMsg"
          variant="neutral"
          data-testid="visit-report-msg"
          role="alert"
        >
          {{ reportMsg }}
        </ProBadge>
      </div>
      <div class="pro-visit-report__footer-actions">
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
        @click="saveVisitReport('save')"
      >
        {{ $t('calendar.saveReport') }}
      </ProButton>
      </div>
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
import { canonicalizeReportMarkdown } from '~/utils/reportRichText'
import { probeAudioDurationSec } from '~/utils/audioDuration'
import { useActiveConsultation } from '~/composables/useActiveConsultation'
import {
  isAdvancedImproveControlStep,
  type AdvancedImproveState,
} from '~/utils/advancedImproveStatus'
import {
  copyVisitReportMarkdown,
  downloadVisitReportMarkdown,
  normalizeReportCitations,
  openVisitReportPdfBlob,
} from '~/utils/visit-report-export'

export type VisitReportAuthor = {
  id?: string
  authorUserId?: string
  authorFullName?: string
  mine?: boolean
  status?: string
  bodyText?: string
  transcriptText?: string
  improvedText?: string
  isReference?: boolean
}

const props = withDefaults(
  defineProps<{
    visitId: string
    /** When true, hide save/improve/finalize/audio (ACL pets.write_clinical). */
    readonly?: boolean
    /** Visit date shown under the title and used to prefix the report body ("Date du : …"). */
    visitScheduledAt?: string
    /**
     * Stretch notes/CR panes to fill a tall host (consultation `full` modal).
     * Default compact — calendar / history / pet modals must not inherit fill min-heights.
     */
    fillHeight?: boolean
  }>(),
  { fillHeight: false },
)

/**
 * Origine d'un `saved` :
 * - `save` : clic explicite « Enregistrer » (le host peut proposer la suite)
 * - `improve` : PUT technique avant/après IA — le CR reste à l'écran
 * - `flush` : leave-guard / veille — le host gère déjà la fermeture
 */
export type VisitReportSavedOrigin = 'save' | 'improve' | 'flush'

const emit = defineEmits<{
  saved: [origin: VisitReportSavedOrigin]
  finalized: []
  busy: [value: boolean]
}>()

const { t } = useI18n()
const { mapError } = useApiError()
const { formatDate } = useFormatters()
const runtimeConfig = useRuntimeConfig()
const aiCrAdvancedEnabled = computed(() => Boolean(runtimeConfig.public.aiCrAdvancedEnabled))
const { start: startAdvancedImprove, stop: stopAdvancedImprove, cancelRemote: cancelAdvancedImproveRemote } = useAdvancedImproveStream()

const reportBody = ref('')
const reportPersistedBody = ref('')
const reportTranscript = ref('')
const reportPersistedTranscript = ref('')
const reportImproved = ref('')
const reportStatus = ref('')
const reportIsReference = ref(false)
const lastImproveCitations = ref<string[]>([])
const exportStripCitations = ref(true)
const reportBusy = ref(false)
const exportBusy = ref(false)
const saveInFlight = ref(false)
const improveInFlight = ref(false)
const advancedImproveInFlight = ref(false)
const advancedSteps = ref<{ agent: string; label: string; at?: string }[]>([])
/** État SSE `status` du run avancé + phase locale `starting` (avant le 1er événement). */
type AdvancedUiState = 'starting' | AdvancedImproveState | ''
const advancedStatus = ref<AdvancedUiState>('')
const advancedStatusLabel = computed(() => {
  const state = advancedStatus.value
  switch (state) {
    case 'starting': return t('calendar.reportImproveAdvancedStatusStarting')
    case 'crew_warming': return t('calendar.reportImproveAdvancedStatusWarming')
    case 'crew_ready': return t('calendar.reportImproveAdvancedStatusReady')
    case 'running': return t('calendar.reportImproveAdvancedStatusRunning')
    case '': return ''
    default: {
      const exhaustive: never = state
      return exhaustive
    }
  }
})
/** True while Stop→transcribe or file upload transcription is in flight. */
const transcribeInFlight = ref(false)
const qualityBusy = ref(false)
const referenceBusy = ref(false)
const showAiQualityBar = ref(false)
const targetLocale = ref('auto')
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

const dictating = ref(false)
const dictationSeconds = ref(0)
const MAX_DICTATION_SECONDS = 10 * 60
let dictationTimer: ReturnType<typeof setInterval> | null = null
let mediaRecorder: MediaRecorder | null = null
let mediaStream: MediaStream | null = null
let recordChunks: Blob[] = []
/** In-flight Stop → transcribe; flush waits on this to avoid overwriting the transcript. */
let stopDictationInFlight: Promise<void> | null = null

// flush sync : le busy(false) final doit partir AVANT un éventuel démontage du
// panneau (emit('saved') → le parent bascule d'écran) — sinon le parent reste
// bloqué busy et waitUntilReportIdle ne se résout jamais.
watch(reportBusy, (busy) => {
  emit('busy', busy)
}, { immediate: true, flush: 'sync' })

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
  canonicalizeReportMarkdown(reportBody.value) !== canonicalizeReportMarkdown(reportPersistedBody.value)
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
  reportIsReference.value = mapped.isReference
  lastImproveCitations.value = normalizeReportCitations(data?.lastImproveCitations)
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
    isReference: author.isReference,
  })
  reportBody.value = mapped.bodyText
  reportPersistedBody.value = mapped.bodyText
  reportTranscript.value = mapped.transcriptText
  reportPersistedTranscript.value = mapped.transcriptText
  reportImproved.value = mapped.improvedText
  reportStatus.value = mapped.status
  reportIsReference.value = mapped.isReference
  lastImproveCitations.value = []
}

async function copyReportMarkdown() {
  if (!reportBody.value.trim() || exportBusy.value) return
  try {
    await copyVisitReportMarkdown(canonicalizeReportMarkdown(reportBody.value), {
      stripCitations: exportStripCitations.value,
    })
    reportMsg.value = t('calendar.reportCopiedMd')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  }
}

function downloadReportMarkdown() {
  if (!reportBody.value.trim() || exportBusy.value) return
  try {
    downloadVisitReportMarkdown(canonicalizeReportMarkdown(reportBody.value), 'cr', {
      stripCitations: exportStripCitations.value,
    })
    reportMsg.value = t('calendar.reportDownloadedMd')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  }
}

async function downloadReportPdf() {
  if (!props.visitId || !reportBody.value.trim() || exportBusy.value) return
  exportBusy.value = true
  reportMsg.value = ''
  try {
    // Persist current body so PDF matches the editor (attachment uses stored report).
    if (dirty.value && !reportLocked.value && !props.readonly && !viewingPeerReport.value) {
      await $fetch(`/api/visits/${props.visitId}/report`, {
        method: 'PUT',
        body: {
          bodyText: normalizeReportText(reportBody.value),
          transcriptText: normalizeReportText(reportTranscript.value),
        },
      })
      reportPersistedBody.value = reportBody.value
      reportPersistedTranscript.value = reportTranscript.value
    }
    const mode = await openVisitReportPdfBlob(props.visitId, {
      stripCitations: exportStripCitations.value,
    })
    reportMsg.value = mode === 'download'
      ? t('calendar.reportDownloadedPdf')
      : t('calendar.reportOpenedPdf')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    exportBusy.value = false
  }
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
  reportIsReference.value = false
  showAiQualityBar.value = false
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

async function saveVisitReport(origin: VisitReportSavedOrigin = 'save'): Promise<boolean> {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return false
  // Garde-fou : un handler @click mal câblé passerait l'événement DOM en 1er argument.
  const savedOrigin: VisitReportSavedOrigin = origin === 'improve' || origin === 'flush' ? origin : 'save'
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
    emit('saved', savedOrigin)
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
  return saveVisitReport('flush')
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
  return saveVisitReport('flush')
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
  if (advancedImproveInFlight.value) return
  const sourceRaw = reportTranscript.value.trim() || reportBody.value.trim()
  if (!sourceRaw) return
  applyDatePrefixIfNeeded()
  reportBusy.value = true
  improveInFlight.value = true
  reportMsg.value = ''
  let putSucceeded = false
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
    // Do NOT emit('saved') here — that would unlock Traitements (DAF) before the IA text arrives.
    // Note : les `saved` de ce flux portent l'origine `improve` — le host garde le CR
    // affiché pour que le véto relise la proposition IA (pas de bascule vers le hub).
    applyReportPayload(putRes.data ?? putRes)
    putSucceeded = true
    // Flat BFF path: nested …/report/improve is registered but not matched by rou3
    // when …/report (GET/PUT) is also a leaf — see 03d-visit-report-ai-bff.
    const res: any = await $fetch(`/api/visits/${props.visitId}/report-improve`, {
      method: 'POST',
      body: {
        sourceText: normalizeReportText(sourceRaw),
        targetLocale: targetLocale.value || 'auto',
      },
    })
    applyReportPayload(res.data ?? res)
    showAiQualityBar.value = true
    reportMsg.value = t('calendar.reportImproved')
    void loadVisitReports(props.visitId)
    emit('saved', 'improve')
  } catch (e: any) {
    reportMsg.value = mapError(e)
    // PUT already persisted — mark saved so leave-guard / Enregistrer stay coherent (DAF OK post-flight).
    if (putSucceeded) emit('saved', 'improve')
  } finally {
    improveInFlight.value = false
    reportBusy.value = false
  }
}

async function improveVisitReportAdvanced() {
  if (!aiCrAdvancedEnabled.value) return
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return
  if (improveInFlight.value || advancedImproveInFlight.value) return
  const sourceRaw = reportTranscript.value.trim() || reportBody.value.trim()
  if (!sourceRaw) return
  applyDatePrefixIfNeeded()
  reportBusy.value = true
  advancedImproveInFlight.value = true
  advancedSteps.value = []
  advancedStatus.value = 'starting'
  reportMsg.value = ''
  let putSucceeded = false
  try {
    const putRes: any = await $fetch(`/api/visits/${props.visitId}/report`, {
      method: 'PUT',
      body: {
        bodyText: normalizeReportText(reportBody.value),
        transcriptText: normalizeReportText(reportTranscript.value),
      },
    })
    applyReportPayload(putRes.data ?? putRes)
    putSucceeded = true
    let finalReport = ''
    await startAdvancedImprove(
      props.visitId,
      {
        sourceText: normalizeReportText(sourceRaw),
        targetLocale: targetLocale.value || 'auto',
      },
      {
        onStep: (step) => {
          // Warm-up / ready / running : ligne d'état seulement (pas la liste agents).
          if (isAdvancedImproveControlStep(step)) return
          if (step?.agent && step?.label) advancedSteps.value = [...advancedSteps.value, step]
        },
        onStatus: (state) => { advancedStatus.value = state },
        onFinal: (report) => { finalReport = report || '' },
        onError: (code) => {
          if (code === 'cancelled') {
            reportMsg.value = t('calendar.reportImproveAdvancedCancelled')
          }
          else if (code === 'crewai_unavailable') {
            reportMsg.value = t('calendar.reportImproveAdvancedUnavailable')
          }
          else {
            reportMsg.value = mapError({ data: { error: code } })
          }
        },
      },
    )
    if (finalReport.trim()) {
      reportBody.value = finalReport
      reportImproved.value = finalReport
      showAiQualityBar.value = true
      reportMsg.value = t('calendar.reportImprovedAdvanced')
      // Re-fetch to sync persisted improvedText from API.
      const getRes: any = await $fetch(`/api/visits/${props.visitId}/report`)
      applyReportPayload(getRes.data ?? getRes)
      void loadVisitReports(props.visitId)
      emit('saved', 'improve')
    } else if (!reportMsg.value) {
      reportMsg.value = t('calendar.reportImproveAdvancedFailed')
      if (putSucceeded) emit('saved', 'improve')
    }
  } catch (e: any) {
    reportMsg.value = mapError(e)
    if (putSucceeded) emit('saved', 'improve')
  } finally {
    stopAdvancedImprove()
    advancedImproveInFlight.value = false
    advancedStatus.value = ''
    reportBusy.value = false
  }
}

async function cancelAdvancedImprove() {
  if (!advancedImproveInFlight.value) return
  await cancelAdvancedImproveRemote()
}

async function submitAiQuality(good: boolean) {
  if (qualityBusy.value) return
  qualityBusy.value = true
  try {
    await $fetch('/api/me/ai-module/feedback', {
      method: 'POST',
      body: {
        nps: good ? 9 : 3,
        source: 'visit_report',
        comment: good ? 'cr_quality_ok' : 'cr_quality_needs_work',
        frictionTags: good ? [] : ['cr_quality'],
      },
    })
    showAiQualityBar.value = false
    reportMsg.value = t('calendar.reportQualityThanks')
  } catch (e: any) {
    showAiQualityBar.value = false
    const status = e?.statusCode ?? e?.status ?? e?.data?.statusCode ?? e?.response?.status
    // Module absent / unpaid: don't overwrite the successful improve message.
    if (status === 404 || status === 403 || status === 402) {
      reportMsg.value = t('calendar.reportImproved')
    } else {
      reportMsg.value = mapError(e)
    }
  } finally {
    qualityBusy.value = false
  }
}

async function onReferenceToggle() {
  if (referenceBusy.value || reportStatus.value !== 'final') return
  const checked = reportIsReference.value
  referenceBusy.value = true
  reportMsg.value = ''
  try {
    const res: any = await $fetch(`/api/visits/${props.visitId}/report-reference`, {
      method: 'PATCH',
      body: { isReference: checked },
    })
    applyReportPayload(res.data ?? res)
    reportMsg.value = checked
      ? t('calendar.reportReferenceOn')
      : t('calendar.reportReferenceOff')
  } catch (e: any) {
    reportIsReference.value = !checked
    reportMsg.value = mapError(e)
  } finally {
    referenceBusy.value = false
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
  transcribeInFlight.value = true
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
    transcribeInFlight.value = false
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
    // Show Arrêter loading before the recording banner swaps to the transcription banner.
    transcribeInFlight.value = true
    reportBusy.value = true
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
      transcribeInFlight.value = false
      reportBusy.value = false
      return
    }
    const ext = blob.type.includes('mp4') ? 'm4a' : 'webm'
    // Prefer live clock; fall back to blob metadata (often missing on MediaRecorder webm).
    let durationSec = recordedSec
    if (durationSec <= 0) {
      durationSec = await probeAudioDurationSec(blob)
    }
    // transcribeAudio owns reportBusy / transcribeInFlight from here.
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
  if (advancedImproveInFlight.value) {
    void cancelAdvancedImproveRemote()
  }
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
    if (prev && advancedImproveInFlight.value) {
      await cancelAdvancedImproveRemote()
    }
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
.pro-visit-report {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  min-height: 0;
}

.pro-visit-report__scroll {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.visit-report-split {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.85rem;
}

@media (min-width: 900px) {
  .visit-report-split {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
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
  min-width: 0;
  min-height: 0;
}

.visit-report-pane__header {
  flex-shrink: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.visit-report-lang__select {
  min-width: 9.5rem;
  padding: 0.35rem 0.5rem;
  font-size: 0.85rem;
}

.visit-report-quality {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.5rem 0.65rem;
  border: 1px dashed var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-bg, #f8fafc);
}

.visit-report-export {
  flex-wrap: wrap;
  margin-top: 0.35rem;
  align-items: center;
}

.visit-report-export__strip {
  margin: 0;
  font-size: 0.85rem;
}

.visit-report-citations {
  margin-top: 0.65rem;
  padding: 0.55rem 0.7rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-bg, #f8fafc);
}

.visit-report-citations__title {
  margin: 0 0 0.35rem;
  font-size: 0.85rem;
  color: var(--pf-vet-primary);
}

.visit-report-citations__list {
  margin: 0;
  padding-left: 1.1rem;
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #6b7280);
}

.visit-report-reference {
  margin: 0.25rem 0 0;
}

.visit-report-pane__header .pro-hint {
  margin: 0.15rem 0 0;
}

.visit-report-pane__title {
  margin: 0;
  font-size: 1rem;
  color: var(--pf-vet-primary);
}

.visit-report-pane__toolbar {
  flex-wrap: wrap;
}

.visit-report-pane__ai-hint {
  flex-shrink: 0;
  margin: 0;
}

.visit-report-pane__textarea {
  min-height: 280px;
  width: 100%;
  resize: vertical;
  flex: 1;
}

.visit-report-recording {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0.9rem;
  border-radius: var(--pf-vet-radius, 8px);
  background: color-mix(in srgb, var(--pf-vet-alert) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--pf-vet-alert) 40%, transparent);
}

.visit-report-recording--busy {
  background: color-mix(in srgb, var(--pf-vet-accent) 12%, transparent);
  border-color: color-mix(in srgb, var(--pf-vet-accent) 40%, transparent);
}

.visit-report-recording--busy .visit-report-recording__mic,
.visit-report-recording--busy .visit-report-recording__info {
  color: var(--pf-vet-accent);
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
  flex-shrink: 0;
  flex-wrap: wrap;
}

.pro-visit-report__footer {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0;
  padding: 0.5rem 0 0.15rem;
  border-top: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface, #fff);
}

.visit-report-status-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
  min-width: 0;
  flex: 1 1 auto;
}

.visit-report-status-tags :deep(.pro-badge) {
  font-size: 0.7rem;
  font-weight: 600;
  line-height: 1.2;
  padding: 0.15rem 0.45rem;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pro-visit-report__footer-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-left: auto;
}

.pro-visit-report__file {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.visit-report-history {
  flex-shrink: 0;
  margin-top: 0.15rem;
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

/*
 * Host modale avec ProModal contain-scroll (ou size=full) :
 * head + footer épinglés, seul __scroll défile.
 * Ne pas activer hors de ces hosts (calendrier embed = flux naturel).
 */
.pro-visit-report--fill {
  flex: 1 1 auto;
  align-self: stretch;
  width: 100%;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.pro-visit-report--fill .pro-visit-report__scroll {
  flex: 1 1 auto;
  min-height: 0;
  /* Scroll plutôt que rognage : les blocs optionnels (historique déplié, barre
     qualité IA, traitements DAF) s'additionnent et écrasaient les panes jusqu'à
     rendre l'éditeur inatteignable. Le scroll n'apparaît que si ça ne rentre pas. */
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  padding-right: 0.15rem;
}

.pro-visit-report--fill .visit-report-split {
  flex: 1 1 auto;
  /* Plancher : les panes gardent une hauteur exploitable même avec historique +
     barre qualité + traitements ouverts (le scroll du conteneur prend le relais). */
  min-height: 14rem;
  overflow: hidden;
  /* Single row that respects the flex-bounded split height (avoids TipTap overflow paint). */
  grid-template-rows: minmax(0, 1fr);
}

.pro-visit-report--fill .visit-report-pane {
  /* Scroll plutôt que rognage : les blocs optionnels empilés sous l'éditeur
     (export, barre qualité, feedback IA, sources citées) s'additionnent. */
  overflow-y: auto;
  min-height: 0;
  height: 100%;
}

.pro-visit-report--fill .visit-report-pane__header .pro-hint {
  display: none;
}

.pro-visit-report--fill .visit-report-recording {
  padding: 0.4rem 0.65rem;
  gap: 0.5rem;
}

.pro-visit-report--fill .visit-report-pane__textarea {
  flex: 1 1 auto;
  min-height: 0;
  resize: none;
}

/* TipTap root (was .pro-md-report before rich editor migration). */
.pro-visit-report--fill :deep(.pro-rich-report) {
  flex: 1 1 auto;
  /* Plancher : les blocs optionnels empilés dans le pane (barre d'export,
     barre qualité, feedback IA, sources citées) écrasaient l'éditeur à 0 px —
     CR inéditable. Le pane ci-dessus prend le scroll à la place. */
  min-height: 8rem;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.pro-visit-report--fill :deep(.pro-rich-report__editor) {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
}

.pro-visit-report--fill :deep(.pro-rich-report__editor .ProseMirror) {
  min-height: 100%;
}

.pro-visit-report--fill .visit-report-history {
  max-height: 2.25rem;
  overflow: hidden;
  padding: 0.35rem 0.65rem;
}

.pro-visit-report--fill .visit-report-history[open] {
  max-height: min(40vh, 16rem);
  overflow: auto;
}

/* Le head est hors zone de scroll et flex-shrink:0 : déplié, « Aide » (image ~36rem)
   prenait toute la hauteur du panneau et rognait les panes — barre d'outils
   « Améliorer » inatteignable, éditeur à hauteur nulle. Même plafond + scroll interne
   que l'historique ci-dessus. */
.pro-visit-report--fill :deep(.visit-report-howto[open]) {
  max-height: min(40vh, 16rem);
  overflow: auto;
}
</style>
