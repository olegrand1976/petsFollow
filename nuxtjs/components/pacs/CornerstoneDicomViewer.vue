<script setup lang="ts">
/**
 * Cornerstone3D stack viewer (P2.1) — DICOM via BFF /file + tools pan/zoom/W/L/Length.
 * Opt-in: NUXT_PUBLIC_PACS_VIEWER_ENGINE=cornerstone (canvas remains default).
 */
import type { PacsDebugEntry } from '~/composables/usePacsDebugLog'
import { pacsDebugFetch } from '~/composables/usePacsDebugLog'
import { parsePixelSpacingMm, type PacsSpacingMm } from '~/utils/pacs-measure'

const props = defineProps<{
  leftInstanceId?: string
  rightInstanceId?: string
  compare?: boolean
  debug?: boolean
}>()

const emit = defineEmits<{
  debug: [entry: PacsDebugEntry]
}>()

type CsTool = 'pan' | 'zoom' | 'wl' | 'measure'

const { t } = useI18n()
const leftEl = ref<HTMLDivElement | null>(null)
const rightEl = ref<HTMLDivElement | null>(null)
const loadError = ref('')
const busy = ref(false)
const metaLabel = ref('')
const voiLabel = ref('')
const tool = ref<CsTool>('wl')
const frameIndex = ref(0)
const maxFrame = ref(0)
const pixelSpacing = ref<PacsSpacingMm | null>(null)
/** True only after calibrateImageSpacing applied from API metadata (aligned with Length). */
const spacingApplied = ref(false)
const measureCalibrated = computed(() => spacingApplied.value)
const downloadBusy = ref(false)
const TOOL_NAMES = ['pan', 'zoom', 'wl', 'measure'] as const

const instanceKey = Math.random().toString(36).slice(2, 10)
const ENGINE_ID = `pf-pacs-cs-engine-${instanceKey}`
const TOOL_GROUP_ID = `pf-pacs-cs-tools-${instanceKey}`
let renderingEngine: any = null
let initPromise: Promise<void> | null = null
let toolsRegistered = false
let voiListenerEl: HTMLDivElement | null = null
let voiHandler: ((ev: Event) => void) | null = null

function onDebug(entry: PacsDebugEntry) {
  if (props.debug) emit('debug', entry)
}

async function apiFetch<T = any>(url: string, opts?: Record<string, any>) {
  if (props.debug) return pacsDebugFetch<T>(url, opts, onDebug)
  return $fetch<T>(url, opts)
}

async function ensureCornerstone() {
  if (initPromise) return initPromise
  initPromise = (async () => {
    const core = await import('@cornerstonejs/core')
    const tools = await import('@cornerstonejs/tools')
    const dicomImageLoader = await import('@cornerstonejs/dicom-image-loader')
    await core.init()
    await tools.init()
    dicomImageLoader.init({ maxWebWorkers: 1 })
    if (!toolsRegistered) {
      tools.addTool(tools.PanTool)
      tools.addTool(tools.ZoomTool)
      tools.addTool(tools.WindowLevelTool)
      tools.addTool(tools.LengthTool)
      tools.addTool(tools.StackScrollTool)
      toolsRegistered = true
    }
  })()
  return initPromise
}

function imageIdFor(instanceId: string, frame = 0) {
  const path = `/api/pacs/instances/${encodeURIComponent(instanceId)}/file`
  const base = `wadouri:${window.location.origin}${path}`
  return frame > 0 ? `${base}?frame=${frame}` : base
}

function stackIds(instanceId: string, frames: number) {
  const n = Math.max(1, frames)
  return Array.from({ length: n }, (_, i) => imageIdFor(instanceId, i))
}

async function loadMetadata(instanceId: string) {
  try {
    const res: any = await apiFetch(`/api/pacs/instances/${instanceId}/metadata`)
    const data = res.data ?? res
    pixelSpacing.value = parsePixelSpacingMm(data.pixelSpacingMm)
    spacingApplied.value = false
    const spacing = pixelSpacing.value ? pixelSpacing.value.join('×') : null
    const nFrames = Number(data.numberOfFrames)
    maxFrame.value = Number.isFinite(nFrames) && nFrames > 0 ? nFrames - 1 : 0
    const parts = [
      data.modality,
      spacing ? `${spacing} mm` : null,
      maxFrame.value > 0 ? `${maxFrame.value + 1} frames` : null,
    ].filter(Boolean)
    metaLabel.value = parts.join(' · ')
  } catch {
    metaLabel.value = ''
    maxFrame.value = 0
    pixelSpacing.value = null
    spacingApplied.value = false
  }
}

async function applyApiSpacingCalibration(imageIds: string[]) {
  spacingApplied.value = false
  if (!renderingEngine || !pixelSpacing.value || !imageIds.length) return
  try {
    const tools = await import('@cornerstonejs/tools')
    const core = await import('@cornerstonejs/core')
    const [rowMm, colMm] = pixelSpacing.value
    const calibration = {
      type: core.Enums.CalibrationTypes.USER,
      rowPixelSpacing: rowMm,
      columnPixelSpacing: colMm,
    }
    for (const imageId of imageIds) {
      tools.utilities.calibrateImageSpacing(imageId, renderingEngine, calibration)
    }
    spacingApplied.value = true
  } catch {
    // Viewer stays usable; Length stays in px until calibration succeeds.
    spacingApplied.value = false
  }
}

function destroyToolGroup() {
  return import('@cornerstonejs/tools').then((tools) => {
    try {
      tools.ToolGroupManager.destroyToolGroup(TOOL_GROUP_ID)
    } catch { /* ignore */ }
  })
}

async function setupToolGroup(viewportIds: string[]) {
  const tools = await import('@cornerstonejs/tools')
  await destroyToolGroup()
  const group = tools.ToolGroupManager.createToolGroup(TOOL_GROUP_ID)
  if (!group) return
  group.addTool(tools.PanTool.toolName)
  group.addTool(tools.ZoomTool.toolName)
  group.addTool(tools.WindowLevelTool.toolName)
  group.addTool(tools.LengthTool.toolName)
  group.addTool(tools.StackScrollTool.toolName)
  for (const id of viewportIds) {
    group.addViewport(id, ENGINE_ID)
  }
  // Wheel always scrolls the stack; middle/right keep pan/zoom as secondary.
  group.setToolActive(tools.StackScrollTool.toolName, {
    bindings: [{ mouseButton: tools.Enums.MouseBindings.Wheel }],
  })
  group.setToolActive(tools.PanTool.toolName, {
    bindings: [{ mouseButton: tools.Enums.MouseBindings.Auxiliary }],
  })
  group.setToolActive(tools.ZoomTool.toolName, {
    bindings: [{ mouseButton: tools.Enums.MouseBindings.Secondary }],
  })
  await applyPrimaryTool(tool.value)
}

async function applyPrimaryTool(next: CsTool) {
  const tools = await import('@cornerstonejs/tools')
  const group = tools.ToolGroupManager.getToolGroup(TOOL_GROUP_ID)
  if (!group) return
  const primary = tools.Enums.MouseBindings.Primary
  const map: Record<CsTool, string> = {
    pan: tools.PanTool.toolName,
    zoom: tools.ZoomTool.toolName,
    wl: tools.WindowLevelTool.toolName,
    measure: tools.LengthTool.toolName,
  }
  for (const name of Object.values(map)) {
    try {
      group.setToolPassive(name)
    } catch { /* not added yet */ }
  }
  group.setToolActive(map[next], {
    bindings: [{ mouseButton: primary }],
  })
}

function updateVoiFromViewport() {
  if (!renderingEngine) {
    voiLabel.value = ''
    return
  }
  try {
    const vp = renderingEngine.getViewport('left')
    const range = vp?.getProperties?.()?.voiRange
    if (!range || range.upper == null || range.lower == null) {
      voiLabel.value = ''
      return
    }
    const ww = Math.round(range.upper - range.lower)
    const wc = Math.round((range.upper + range.lower) / 2)
    voiLabel.value = `WW ${ww} · WC ${wc}`
  } catch {
    voiLabel.value = ''
  }
}

async function attachVoiListener(el: HTMLDivElement | null) {
  const core = await import('@cornerstonejs/core')
  if (voiListenerEl && voiHandler) {
    voiListenerEl.removeEventListener(core.Enums.Events.VOI_MODIFIED, voiHandler)
    voiListenerEl.removeEventListener(core.Enums.Events.IMAGE_RENDERED, voiHandler)
    voiListenerEl.removeEventListener(core.Enums.Events.STACK_NEW_IMAGE, voiHandler)
  }
  voiListenerEl = el
  if (!el) {
    voiHandler = null
    return
  }
  voiHandler = (ev: Event) => {
    const detail = (ev as CustomEvent)?.detail
    if (detail && typeof detail.imageIdIndex === 'number') {
      frameIndex.value = detail.imageIdIndex
    }
    updateVoiFromViewport()
  }
  el.addEventListener(core.Enums.Events.VOI_MODIFIED, voiHandler)
  el.addEventListener(core.Enums.Events.IMAGE_RENDERED, voiHandler)
  el.addEventListener(core.Enums.Events.STACK_NEW_IMAGE, voiHandler)
}

async function bindStack(element: HTMLDivElement | null, instanceId: string | undefined, viewportId: string) {
  if (!element || !instanceId || !renderingEngine) return
  const core = await import('@cornerstonejs/core')
  element.innerHTML = ''
  renderingEngine.enableElement({
    viewportId,
    element,
    type: core.Enums.ViewportType.STACK,
  })
  const vp = renderingEngine.getViewport(viewportId)
  const ids = stackIds(instanceId, maxFrame.value + 1)
  await vp.setStack(ids, Math.min(frameIndex.value, ids.length - 1))
  await applyApiSpacingCalibration(ids)
  vp.render()
  if (viewportId === 'left') {
    await attachVoiListener(element)
    updateVoiFromViewport()
  }
}

async function reload() {
  if (!props.leftInstanceId) return
  busy.value = true
  loadError.value = ''
  frameIndex.value = 0
  try {
    await ensureCornerstone()
    const core = await import('@cornerstonejs/core')
    await destroyToolGroup()
    if (renderingEngine) {
      try { renderingEngine.destroy() } catch { /* ignore */ }
      renderingEngine = null
    }
    renderingEngine = new core.RenderingEngine(ENGINE_ID)
    await loadMetadata(props.leftInstanceId)
    await nextTick()
    const viewportIds = ['left']
    await bindStack(leftEl.value, props.leftInstanceId, 'left')
    if (props.compare && props.rightInstanceId) {
      viewportIds.push('right')
      await bindStack(rightEl.value, props.rightInstanceId, 'right')
    }
    await setupToolGroup(viewportIds)
    updateVoiFromViewport()
  } catch (e: any) {
    loadError.value = fetchCsError(e) || e?.message || t('pacs.previewError')
  } finally {
    busy.value = false
  }
}

function fetchCsError(e: unknown): string {
  const err = e as { data?: any }
  const d = err?.data
  const msg = d?.message || d?.error?.message
  if (typeof msg === 'string' && msg && msg !== 'Error') return msg
  const key = d?.msgKey || d?.error?.msgKey
  if (key === 'pacs_instance_unavailable') return t('pacs.instanceUnavailable')
  return ''
}

async function setTool(next: CsTool) {
  tool.value = next
  await applyPrimaryTool(next)
}

function looksLikeDicom(buf: Uint8Array): boolean {
  return buf.length >= 132
    && buf[128] === 0x44 && buf[129] === 0x49 && buf[130] === 0x43 && buf[131] === 0x4d
}

async function downloadDicom() {
  const id = props.leftInstanceId
  if (!id || downloadBusy.value) return
  downloadBusy.value = true
  try {
    const blob = await apiFetch<Blob>(`/api/pacs/instances/${id}/file`, { responseType: 'blob' })
    const head = new Uint8Array(await blob.slice(0, 132).arrayBuffer())
    if (!looksLikeDicom(head)) throw new Error('not_dicom')
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${id.slice(0, 12)}.dcm`
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    loadError.value = t('pacs.downloadError')
  } finally {
    downloadBusy.value = false
  }
}

async function stepFrame(delta: number) {
  if (!renderingEngine || !props.leftInstanceId) return
  const vp = renderingEngine.getViewport('left')
  if (!vp) return
  const next = Math.max(0, Math.min(maxFrame.value, frameIndex.value + delta))
  if (next === frameIndex.value) return
  frameIndex.value = next
  try {
    await vp.setImageIdIndex(next)
    vp.render()
    updateVoiFromViewport()
  } catch {
    /* ignore */
  }
}

watch(
  () => [props.leftInstanceId, props.rightInstanceId, props.compare] as const,
  () => { void nextTick(() => void reload()) },
)

onMounted(() => { void reload() })

onBeforeUnmount(async () => {
  if (voiListenerEl && voiHandler) {
    try {
      const core = await import('@cornerstonejs/core')
      voiListenerEl.removeEventListener(core.Enums.Events.VOI_MODIFIED, voiHandler)
      voiListenerEl.removeEventListener(core.Enums.Events.IMAGE_RENDERED, voiHandler)
      voiListenerEl.removeEventListener(core.Enums.Events.STACK_NEW_IMAGE, voiHandler)
    } catch { /* ignore */ }
  }
  voiListenerEl = null
  voiHandler = null
  await destroyToolGroup()
  if (renderingEngine) {
    try { renderingEngine.destroy() } catch { /* ignore */ }
    renderingEngine = null
  }
})
</script>

<template>
  <div class="cs-viewer" data-testid="dicom-viewer-cornerstone">
    <p v-if="loadError" class="pro-inline-feedback pro-inline-feedback--error" role="alert" data-testid="dicom-preview-error">
      {{ loadError }}
    </p>
    <div class="cs-viewer__toolbar" role="toolbar">
      <span class="cs-viewer__badge" data-testid="dicom-engine-badge">Cornerstone3D</span>
      <button
        v-for="tb in TOOL_NAMES"
        :key="tb"
        type="button"
        class="cs-viewer__tool"
        :class="{ 'is-active': tool === tb }"
        :data-testid="`dicom-tool-${tb}`"
        @click="setTool(tb)"
      >
        {{ t(`pacs.tools.${tb}`) }}
      </button>
      <button
        type="button"
        class="cs-viewer__tool"
        data-testid="dicom-tool-frame-prev"
        :disabled="frameIndex <= 0"
        @click="stepFrame(-1)"
      >
        {{ t('pacs.tools.prevFrame') }}
      </button>
      <span class="cs-viewer__frame" data-testid="dicom-frame-label">
        {{ t('pacs.tools.frame', { n: frameIndex + 1 }) }}
      </span>
      <button
        type="button"
        class="cs-viewer__tool"
        data-testid="dicom-tool-frame-next"
        :disabled="frameIndex >= maxFrame"
        @click="stepFrame(1)"
      >
        {{ t('pacs.tools.nextFrame') }}
      </button>
      <button
        type="button"
        class="cs-viewer__tool"
        data-testid="dicom-tool-download"
        :disabled="!leftInstanceId || downloadBusy"
        @click="downloadDicom"
      >
        {{ t('pacs.tools.download') }}
      </button>
      <span v-if="voiLabel" class="cs-viewer__voi" data-testid="dicom-voi-label">{{ voiLabel }}</span>
      <span
        class="cs-viewer__calib"
        :data-calibrated="measureCalibrated ? '1' : '0'"
        data-testid="dicom-calib-badge"
      >
        {{ measureCalibrated ? t('pacs.tools.measureCalibrated') : t('pacs.tools.measureUncalibrated') }}
      </span>
      <span v-if="metaLabel" class="cs-viewer__meta" data-testid="dicom-meta-label">{{ metaLabel }}</span>
      <span v-if="busy" class="cs-viewer__busy">{{ t('pacs.launch.waking') }}</span>
    </div>
    <p
      v-if="tool === 'measure' && !measureCalibrated"
      class="cs-viewer__hint"
      data-testid="dicom-measure-hint"
    >
      {{ t('pacs.tools.measureHintPx') }}
    </p>
    <div class="cs-viewer__panes" :class="{ 'cs-viewer__panes--compare': compare }">
      <div ref="leftEl" class="cs-viewer__pane" data-testid="dicom-cs-left" />
      <div v-if="compare" ref="rightEl" class="cs-viewer__pane" data-testid="dicom-cs-right" />
    </div>
  </div>
</template>

<style scoped>
.cs-viewer {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 360px;
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  padding: 0.75rem;
}
.cs-viewer__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
  font-size: 0.85rem;
}
.cs-viewer__badge {
  border: 1px solid var(--pf-vet-accent);
  color: var(--pf-vet-accent);
  border-radius: 6px;
  padding: 0.15rem 0.5rem;
  font-weight: 600;
}
.cs-viewer__tool {
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-bg);
  color: var(--pf-vet-primary);
  border-radius: 6px;
  padding: 0.25rem 0.55rem;
  cursor: pointer;
  font: inherit;
}
.cs-viewer__tool.is-active {
  border-color: var(--pf-vet-accent);
  background: color-mix(in srgb, var(--pf-vet-accent) 12%, transparent);
}
.cs-viewer__tool:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.cs-viewer__meta,
.cs-viewer__voi,
.cs-viewer__frame { opacity: 0.8; color: var(--pf-vet-primary); }
.cs-viewer__voi { font-variant-numeric: tabular-nums; font-weight: 600; }
.cs-viewer__calib {
  font-size: 0.78rem;
  opacity: 0.85;
  border: 1px dashed var(--pf-vet-border);
  border-radius: 6px;
  padding: 0.15rem 0.45rem;
  color: var(--pf-vet-primary);
}
.cs-viewer__calib[data-calibrated='1'] {
  border-color: var(--pf-vet-accent);
  color: var(--pf-vet-accent);
}
.cs-viewer__hint {
  margin: 0;
  font-size: 0.8rem;
  opacity: 0.75;
  color: var(--pf-vet-primary);
}
.cs-viewer__busy { opacity: 0.6; }
.cs-viewer__panes {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.5rem;
  min-height: 320px;
  flex: 1;
}
.cs-viewer__panes--compare { grid-template-columns: 1fr 1fr; }
.cs-viewer__pane {
  min-height: 320px;
  background: #0b1220;
  border-radius: 8px;
  overflow: hidden;
  touch-action: none;
}
@media (max-width: 900px) {
  .cs-viewer__panes--compare { grid-template-columns: 1fr; }
}
</style>
