<script setup lang="ts">
import { isPublicFlagOn } from '~/utils/public-feature-flag'
import { resolvePacsViewerEngine } from '~/utils/pacs-viewer-engine'
import type { PacsDebugEntry } from '~/composables/usePacsDebugLog'
import { pacsDebugFetch } from '~/composables/usePacsDebugLog'

const props = defineProps<{
  petId: string
  /** When true, emit fetch traces for admin playground debug. */
  debug?: boolean
  /** Auto-open the first study once the list is loaded (admin playground). */
  autoOpenFirstStudy?: boolean
}>()

const emit = defineEmits<{
  debug: [entry: PacsDebugEntry]
}>()

const { t } = useI18n()
const runtimeConfig = useRuntimeConfig()
const pacsOn = computed(() => isPublicFlagOn(runtimeConfig.public.pacsEnabled))
const viewerEngine = computed(() => resolvePacsViewerEngine(runtimeConfig.public.pacsViewerEngine))
const useCornerstone = computed(() => viewerEngine.value === 'cornerstone')

function onDebug(entry: PacsDebugEntry) {
  if (props.debug) emit('debug', entry)
}

async function apiFetch<T = any>(url: string, opts?: Record<string, any>) {
  if (props.debug) return pacsDebugFetch<T>(url, opts, onDebug)
  return $fetch<T>(url, opts)
}

const { status, loading, waking, hasPolled, error, wake, wakeUntilReady, refresh } = usePacsStatus({
  enabled: pacsOn,
  onDebug: props.debug ? onDebug : undefined,
})

const studies = ref<any[]>([])
const studiesError = ref('')
const uploadBusy = ref(false)
const uploadError = ref('')
const compare = ref(false)
const leftInstanceId = ref('')
const rightInstanceId = ref('')
const selectedStudyId = ref('')
const seriesIds = ref<string[]>([])
const selectedSeriesId = ref('')
const instances = ref<string[]>([])
const fileInput = ref<HTMLInputElement | null>(null)

/** Hold launch pad ~1.4s after ready so step « ready » is visible. */
const celebrateReady = ref(false)
let celebrateTimer: ReturnType<typeof setTimeout> | null = null

watch(
  () => status.value.state,
  (state, prev) => {
    if (state === 'ready' && (prev === 'starting' || prev === 'offline')) {
      celebrateReady.value = true
      if (celebrateTimer) clearTimeout(celebrateTimer)
      celebrateTimer = setTimeout(() => {
        celebrateReady.value = false
        celebrateTimer = null
      }, 1400)
      return
    }
    if (state !== 'ready') {
      celebrateReady.value = false
      if (celebrateTimer) {
        clearTimeout(celebrateTimer)
        celebrateTimer = null
      }
    }
  },
)

onBeforeUnmount(() => {
  if (celebrateTimer) clearTimeout(celebrateTimer)
})

const isStarting = computed(() => waking.value || status.value.state === 'starting')
/** Only show pad during wake/starting or short ready celebration — never on first silent poll. */
const showLaunchPad = computed(() => isStarting.value || celebrateReady.value)
const actionsLocked = computed(() => status.value.state !== 'ready' || isStarting.value || uploadBusy.value)
const wakeDisabled = computed(() => isStarting.value || (!hasPolled.value && loading.value))

const badgeVariant = computed(() => {
  if (status.value.state === 'ready') return 'success'
  if (status.value.state === 'starting' || waking.value) return 'warning'
  return 'neutral'
})

const stateLabel = computed(() => t(`pacs.state.${status.value.state}`))

/** Prefer seed RX / DX over broken CT indexes (storage miss → 502 on study/preview). */
function pickPreferredStudy(items: any[]): any | null {
  if (!items.length) return null
  const score = (s: any) => {
    const mod = String(s.modality || '').toUpperCase()
    const desc = String(s.description || '').toLowerCase()
    if (/rx|thorax|demo|petsfollow/.test(desc)) return 100
    if (/\b(DX|CR|DR|PX)\b/.test(mod)) return 80
    if (/\bCT\b/.test(mod) || /ct[_ ]?chest/.test(desc)) return 10
    return 40
  }
  return [...items].sort((a, b) => score(b) - score(a))[0]
}

function pacsFetchMessage(e: any, fallback: string): string {
  const key = e?.data?.msgKey || e?.data?.error?.msgKey
  if (key === 'pacs_instance_unavailable') return t('pacs.instanceUnavailable')
  if (key === 'pacs_error' || key === 'pacs_unavailable') return t('pacs.orthancError')
  const msg = e?.data?.message || e?.message
  if (typeof msg === 'string' && msg && msg !== 'Error') return msg
  return fallback
}

async function loadStudies() {
  studiesError.value = ''
  try {
    const res: any = await apiFetch(`/api/pets/${props.petId}/pacs/studies`)
    studies.value = res.data ?? res ?? []
    if (
      props.autoOpenFirstStudy
      && !selectedStudyId.value
      && !leftInstanceId.value
      && status.value.state === 'ready'
      && studies.value.length
      && !actionsLocked.value
    ) {
      const preferred = pickPreferredStudy(studies.value)
      if (preferred) await openStudy(preferred)
    }
  } catch (e: any) {
    studiesError.value = pacsFetchMessage(e, 'load_failed')
  }
}

function syncComparePane() {
  if (!compare.value) {
    rightInstanceId.value = ''
    return
  }
  if (!rightInstanceId.value || !instances.value.includes(rightInstanceId.value)) {
    rightInstanceId.value = instances.value.find((id) => id !== leftInstanceId.value) || instances.value[0] || ''
  }
}

async function loadSeries(seriesId: string) {
  selectedSeriesId.value = seriesId
  leftInstanceId.value = ''
  rightInstanceId.value = ''
  instances.value = []
  const series: any = await apiFetch(`/api/pacs/series/${seriesId}`)
  const data = series.data ?? series
  const ids = (data.Instances as string[]) || []
  instances.value = ids
  if (ids[0]) leftInstanceId.value = ids[0]
  syncComparePane()
}

async function openStudy(study: any) {
  if (actionsLocked.value) return
  selectedStudyId.value = study.orthancStudyId
  leftInstanceId.value = ''
  rightInstanceId.value = ''
  instances.value = []
  seriesIds.value = []
  selectedSeriesId.value = ''
  studiesError.value = ''
  const preferred = study.orthancSeriesId
  try {
    try {
      const meta: any = await apiFetch(`/api/pacs/studies/${study.orthancStudyId}`)
      const data = meta.data ?? meta
      const fromStudy = (data.Series as string[]) || []
      seriesIds.value = preferred && !fromStudy.includes(preferred)
        ? [preferred, ...fromStudy]
        : (fromStudy.length ? fromStudy : (preferred ? [preferred] : []))
    } catch (metaErr: any) {
      // Study proxy 502 (Orthanc/GCS) — still try linked series from DB row.
      if (preferred) {
        seriesIds.value = [preferred]
      } else {
        throw metaErr
      }
    }
    const first = seriesIds.value[0]
    if (!first) {
      studiesError.value = t('pacs.orthancError')
      return
    }
    await loadSeries(first)
  } catch (e: any) {
    studiesError.value = pacsFetchMessage(e, t('pacs.orthancError'))
  }
}

async function onSeriesChange() {
  if (!selectedSeriesId.value || actionsLocked.value) return
  try {
    await loadSeries(selectedSeriesId.value)
  } catch (e: any) {
    studiesError.value = e?.data?.message || e?.message || 'open_failed'
  }
}

watch(compare, () => {
  syncComparePane()
})

async function onUpload(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadBusy.value = true
  uploadError.value = ''
  try {
    if (status.value.state !== 'ready') {
      const ok = await wakeUntilReady(90000)
      if (!ok) {
        uploadError.value = t('pacs.launch.notReady')
        return
      }
    }
    const fd = new FormData()
    fd.append('file', file)
    await apiFetch(`/api/pets/${props.petId}/pacs/studies`, { method: 'POST', body: fd })
    await loadStudies()
  } catch (e: any) {
    uploadError.value = e?.data?.message || e?.message || 'upload_failed'
  } finally {
    uploadBusy.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

watch(() => status.value.state, (s) => {
  if (s === 'ready') void loadStudies()
})

onMounted(() => {
  if (status.value.state === 'ready') void loadStudies()
})
</script>

<template>
  <div v-if="pacsOn" class="pacs-container" data-testid="pacs-viewer-container">
    <div class="pacs-container__header">
      <div class="pacs-container__title-row">
        <h3 class="pacs-container__title">{{ t('pacs.title') }}</h3>
        <ProBadge variant="warning" data-testid="pacs-dev-badge">{{ t('nav.tagDev') }}</ProBadge>
        <ProBadge :variant="badgeVariant" data-testid="pacs-status-badge">
          <span v-if="isStarting" class="pacs-spinner" aria-hidden="true" />
          {{ stateLabel }}
          <template v-if="status.latencyMs != null && status.state === 'ready'">
            · {{ status.latencyMs }} ms
          </template>
        </ProBadge>
      </div>
      <div class="pacs-container__actions">
        <ProButton
          v-if="status.state !== 'ready'"
          data-testid="pacs-wake-btn"
          :disabled="wakeDisabled"
          :loading="isStarting"
          @click="wake"
        >
          {{ isStarting ? t('pacs.launch.waking') : t('pacs.wake') }}
        </ProButton>
        <ProButton
          variant="secondary"
          data-testid="pacs-refresh-btn"
          :disabled="loading || isStarting"
          @click="refresh"
        >
          {{ t('pacs.refresh') }}
        </ProButton>
        <label class="pacs-upload" :class="{ 'is-disabled': actionsLocked }">
          <input
            ref="fileInput"
            type="file"
            accept=".dcm,application/dicom"
            data-testid="pacs-upload-input"
            :disabled="actionsLocked"
            @change="onUpload"
          >
          <span class="pro-btn pro-btn--secondary">{{ t('pacs.upload') }}</span>
        </label>
        <label class="pacs-compare" :class="{ 'is-disabled': actionsLocked }">
          <input
            v-model="compare"
            type="checkbox"
            data-testid="pacs-compare-toggle"
            :disabled="actionsLocked"
          >
          {{ t('pacs.compare') }}
        </label>
      </div>
    </div>

    <p v-if="error || studiesError || uploadError" class="pro-inline-feedback pro-inline-feedback--error" role="alert" data-testid="pacs-error">
      {{ error || studiesError || uploadError }}
    </p>

    <PacsLaunchPad
      v-if="showLaunchPad"
      :state="status.state"
      :waking="waking"
      :celebrating="celebrateReady"
    />

    <div
      v-else-if="status.state !== 'ready'"
      class="pacs-container__hint"
      data-testid="pacs-offline-hint"
    >
      <p>{{ t('pacs.offlineHint') }}</p>
      <p class="pacs-container__hint-lock">{{ t('pacs.launch.actionsLocked') }}</p>
    </div>

    <template v-else>
      <ProEmptyState
        v-if="!studies.length"
        :title="t('pacs.emptyTitle')"
        :description="t('pacs.emptyDescription')"
      />
      <div v-else class="pacs-studies">
        <button
          v-for="s in studies"
          :key="s.id"
          type="button"
          class="pacs-study"
          :class="{ 'is-active': selectedStudyId === s.orthancStudyId }"
          data-testid="pacs-study-item"
          :disabled="actionsLocked"
          @click="openStudy(s)"
        >
          <strong>{{ s.description || s.modality || s.orthancStudyId }}</strong>
          <span>{{ s.modality }} · {{ s.studyInstanceUid || s.orthancStudyId }}</span>
        </button>
      </div>

      <div v-if="seriesIds.length > 1" class="pacs-instances">
        <label>
          {{ t('pacs.seriesPicker') }}
          <select
            v-model="selectedSeriesId"
            data-testid="pacs-series-select"
            :disabled="actionsLocked"
            @change="onSeriesChange"
          >
            <option v-for="(id, i) in seriesIds" :key="id" :value="id">
              {{ t('pacs.seriesOption', { n: i + 1, id: id.slice(0, 12) }) }}
            </option>
          </select>
        </label>
      </div>

      <div v-if="instances.length > 1" class="pacs-instances">
        <label>
          {{ t('pacs.leftInstance') }}
          <select v-model="leftInstanceId" data-testid="pacs-left-instance" :disabled="actionsLocked">
            <option v-for="id in instances" :key="id" :value="id">{{ id.slice(0, 12) }}</option>
          </select>
        </label>
        <label v-if="compare">
          {{ t('pacs.rightInstance') }}
          <select v-model="rightInstanceId" data-testid="pacs-right-instance" :disabled="actionsLocked">
            <option v-for="id in instances" :key="'r-'+id" :value="id">{{ id.slice(0, 12) }}</option>
          </select>
        </label>
      </div>

      <ClientOnly>
        <PacsCornerstoneDicomViewer
          v-if="leftInstanceId && useCornerstone"
          :left-instance-id="leftInstanceId"
          :right-instance-id="compare ? rightInstanceId : undefined"
          :compare="compare"
          :debug="debug"
          @debug="onDebug"
        />
        <PacsDicomViewer
          v-else-if="leftInstanceId"
          :left-instance-id="leftInstanceId"
          :right-instance-id="compare ? rightInstanceId : undefined"
          :compare="compare"
        />
        <p
          v-else-if="selectedStudyId"
          class="pacs-container__hint"
          data-testid="pacs-viewer-waiting"
        >
          {{ t('pacs.viewerWaiting') }}
        </p>
      </ClientOnly>
    </template>
  </div>
</template>

<style scoped>
.pacs-container { display: flex; flex-direction: column; gap: 1rem; }
.pacs-container__header { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 0.75rem; }
.pacs-container__title-row { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
.pacs-container__title { margin: 0; font-size: 1.1rem; color: var(--pf-vet-primary); }
.pacs-container__actions { display: flex; flex-wrap: wrap; gap: 0.5rem; align-items: center; }
.pacs-upload input { position: absolute; width: 1px; height: 1px; opacity: 0; }
.pacs-upload { position: relative; cursor: pointer; }
.pacs-upload.is-disabled,
.pacs-compare.is-disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.pacs-upload.is-disabled { pointer-events: none; }
.pacs-compare { display: flex; align-items: center; gap: 0.35rem; font-size: 0.9rem; }
.pacs-spinner {
  display: inline-block; width: 0.7rem; height: 0.7rem; margin-right: 0.35rem;
  border: 2px solid currentColor; border-right-color: transparent; border-radius: 50%;
  animation: pacs-spin 0.7s linear infinite; vertical-align: -1px;
}
@keyframes pacs-spin { to { transform: rotate(360deg); } }
.pacs-studies { display: flex; flex-direction: column; gap: 0.35rem; }
.pacs-study {
  text-align: left; border: 1px solid var(--pf-vet-border); background: var(--pf-vet-bg);
  border-radius: 8px; padding: 0.6rem 0.75rem; cursor: pointer; display: flex; flex-direction: column; gap: 0.15rem;
}
.pacs-study:disabled { opacity: 0.5; cursor: not-allowed; }
.pacs-study.is-active { border-color: var(--pf-vet-accent); }
.pacs-study span { font-size: 0.8rem; opacity: 0.75; }
.pacs-instances { display: flex; flex-wrap: wrap; gap: 1rem; }
.pacs-instances select { margin-left: 0.35rem; }
.pacs-container__hint { color: var(--pf-vet-primary); opacity: 0.9; }
.pacs-container__hint p { margin: 0 0 0.35rem; }
.pacs-container__hint-lock { font-size: 0.88rem; opacity: 0.75; }
</style>
