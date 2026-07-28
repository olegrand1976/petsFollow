<template>
  <div class="pro-visit-report" data-testid="visit-report-panel">
    <label class="pro-label" for="visit-report-body">{{ $t('calendar.reportTitle') }}</label>
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
    <p v-else-if="!readonly" class="pro-hint" data-testid="visit-report-ai-banner">
      {{ $t('calendar.reportAiProposalBanner') }}
    </p>
    <textarea
      id="visit-report-body"
      v-model="reportBody"
      class="pro-input"
      rows="6"
      data-testid="visit-report-body"
      :placeholder="$t('calendar.reportHint')"
      :disabled="reportBusy || reportLocked"
      :readonly="reportLocked"
    />
    <div v-if="!reportLocked" class="pro-flex-gap pro-visit-report__actions">
      <ProButton
        variant="secondary"
        :disabled="reportBusy || reportStatus === 'final'"
        test-id="visit-report-save"
        @click="saveVisitReport"
      >
        {{ $t('calendar.saveReport') }}
      </ProButton>
      <ProButton
        :disabled="reportBusy || reportStatus === 'final'"
        test-id="visit-report-improve"
        @click="improveVisitReport"
      >
        {{ $t('calendar.improveReport') }}
      </ProButton>
      <ProButton
        variant="secondary"
        :disabled="reportBusy || reportStatus === 'final' || !reportBody.trim()"
        test-id="visit-report-finalize"
        @click="finalizeVisitReport"
      >
        {{ $t('calendar.finalizeReport') }}
      </ProButton>
      <label
        v-if="reportStatus !== 'final'"
        class="pro-link-btn pro-visit-report__audio"
      >
        <input
          type="file"
          accept="audio/*,.mp3,.m4a,.wav,.ogg,.webm"
          class="pro-visit-report__file"
          data-testid="visit-report-audio"
          :disabled="reportBusy"
          @change="onReportAudioSelected"
        >
        {{ $t('calendar.transcribeAudio') }}
      </label>
    </div>
    <p v-if="reportStatus === 'final'" class="pro-hint">{{ $t('calendar.reportFinal') }}</p>
    <p v-if="reportMsg" class="pro-hint" data-testid="visit-report-msg">{{ reportMsg }}</p>
    <details
      v-if="reportTranscript || reportImproved || reportHistorySaved"
      class="visit-report-history"
      data-testid="visit-report-history"
    >
      <summary>{{ $t('calendar.reportHistoryTitle') }}</summary>
      <div class="visit-report-history__block">
        <h4>{{ $t('calendar.reportHistoryTranscript') }}</h4>
        <pre v-if="reportTranscript" class="visit-report-history__text">{{ reportTranscript }}</pre>
        <p v-else class="pro-hint">{{ $t('calendar.reportHistoryEmpty') }}</p>
      </div>
      <div class="visit-report-history__block">
        <h4>{{ $t('calendar.reportHistoryImproved') }}</h4>
        <pre v-if="reportImproved" class="visit-report-history__text">{{ reportImproved }}</pre>
        <p v-else class="pro-hint">{{ $t('calendar.reportHistoryEmpty') }}</p>
      </div>
      <div v-if="reportHistorySaved" class="visit-report-history__block">
        <h4>{{ $t('calendar.reportHistorySaved') }}</h4>
        <pre class="visit-report-history__text">{{ reportHistorySaved }}</pre>
      </div>
    </details>

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
}>()

const emit = defineEmits<{
  saved: []
  finalized: []
}>()

const { t } = useI18n()
const { mapError } = useApiError()

const reportBody = ref('')
const reportPersistedBody = ref('')
const reportTranscript = ref('')
const reportImproved = ref('')
const reportStatus = ref('')
const reportBusy = ref(false)
const reportMsg = ref('')
const reportAuthors = ref<VisitReportAuthor[]>([])
const selectedReportAuthorId = ref('')
const audioConsentOpen = ref(false)
const audioConsentChecked = ref(false)
const pendingAudioFile = ref<File | null>(null)

const viewingPeerReport = computed(() => {
  if (!selectedReportAuthorId.value || reportAuthors.value.length === 0) return false
  const selected = reportAuthors.value.find(a => a.authorUserId === selectedReportAuthorId.value)
  return Boolean(selected && !selected.mine)
})

const reportLocked = computed(() =>
  Boolean(props.readonly) || viewingPeerReport.value || reportStatus.value === 'final',
)

const reportHistorySaved = computed(() =>
  persistedHistoryBody(reportPersistedBody.value, reportTranscript.value, reportImproved.value),
)

function applyReportPayload(data: Record<string, unknown> | null | undefined) {
  const mapped = mapVisitReportFields(data)
  reportBody.value = mapped.bodyText
  reportPersistedBody.value = mapped.bodyText
  reportTranscript.value = mapped.transcriptText
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
  reportBody.value = author.bodyText || ''
  reportPersistedBody.value = author.bodyText || ''
  reportTranscript.value = author.transcriptText || ''
  reportImproved.value = author.improvedText || ''
  reportStatus.value = author.status || ''
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
  reportBody.value = ''
  reportPersistedBody.value = ''
  reportTranscript.value = ''
  reportImproved.value = ''
  reportStatus.value = ''
  reportMsg.value = ''
  reportAuthors.value = []
  selectedReportAuthorId.value = ''

  let minePayload: Record<string, unknown> | null = null
  try {
    const res: any = await $fetch(`/api/visits/${visitId}/report`)
    minePayload = (res.data ?? res) as Record<string, unknown>
  } catch {
    minePayload = null
  }

  await loadVisitReports(visitId)

  const mine = reportAuthors.value.find(a => a.mine)
  const peerWithContent = reportAuthors.value.find(a => !a.mine && reportHasContent(a))

  if (mine && reportHasContent(mine)) {
    selectedReportAuthorId.value = mine.authorUserId || ''
    applyReportPayload(minePayload)
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

function selectReportAuthor(author: VisitReportAuthor) {
  selectedReportAuthorId.value = author.authorUserId || ''
  if (author.mine) {
    void loadVisitReport(props.visitId)
    return
  }
  applyPeerReport(author)
}

async function saveVisitReport() {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    const res: any = await $fetch(`/api/visits/${props.visitId}/report`, {
      method: 'PUT',
      body: { bodyText: reportBody.value },
    })
    applyReportPayload(res.data ?? res)
    reportMsg.value = t('calendar.reportSaved')
    void loadVisitReports(props.visitId)
    emit('saved')
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

async function improveVisitReport() {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    await $fetch(`/api/visits/${props.visitId}/report`, {
      method: 'PUT',
      body: { bodyText: reportBody.value },
    })
    const res: any = await $fetch(`/api/visits/${props.visitId}/report/improve`, {
      method: 'POST',
    })
    applyReportPayload(res.data ?? res)
    reportMsg.value = t('calendar.reportImproved')
    void loadVisitReports(props.visitId)
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

async function finalizeVisitReport() {
  if (props.readonly || viewingPeerReport.value || reportStatus.value === 'final') return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    await $fetch(`/api/visits/${props.visitId}/report`, {
      method: 'PUT',
      body: { bodyText: reportBody.value },
    })
    const res: any = await $fetch(`/api/visits/${props.visitId}/report/finalize`, {
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
  audioConsentChecked.value = false
  audioConsentOpen.value = true
}

function cancelAudioConsent() {
  pendingAudioFile.value = null
  audioConsentChecked.value = false
  audioConsentOpen.value = false
}

async function acceptAudioConsent() {
  if (!audioConsentChecked.value) return
  const file = pendingAudioFile.value
  pendingAudioFile.value = null
  audioConsentOpen.value = false
  audioConsentChecked.value = false
  if (!file || reportStatus.value === 'final') return
  reportBusy.value = true
  reportMsg.value = ''
  try {
    const form = new FormData()
    form.append('audio', file, file.name)
    form.append('clientAudioConsent', 'true')
    const res: any = await $fetch(`/api/visits/${props.visitId}/report/transcribe`, {
      method: 'POST',
      body: form,
    })
    applyReportPayload(res.data ?? res)
    reportMsg.value = t('calendar.reportTranscribed')
    void loadVisitReports(props.visitId)
  } catch (e: any) {
    reportMsg.value = mapError(e)
  } finally {
    reportBusy.value = false
  }
}

watch(
  () => props.visitId,
  (id) => {
    if (id) void hydrateVisitReports(id)
  },
  { immediate: true },
)
</script>

<style scoped>
.visit-report-authors {
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
}

.pro-visit-report__actions {
  margin-top: 0.5rem;
  flex-wrap: wrap;
}

.pro-visit-report__audio {
  cursor: pointer;
}

.pro-visit-report__file {
  display: none;
}

.visit-report-history {
  margin-top: 0.75rem;
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
  font-family: inherit;
  font-size: 0.85rem;
  line-height: 1.4;
}
</style>
