<script setup lang="ts">
/**
 * Canvas DICOM viewer (client-only).
 * Loads Orthanc frame previews via BFF; tools: zoom, pan, W/L, measure, arrow, fullscreen, dual-pane, frame scroll, download.
 * Annotations stored in image space (survive resize/fullscreen). Measures use PixelSpacing (mm), scaled if preview resized.
 */
import {
  formatPacsLength,
  measureImageSegment,
  parseMatrixSize,
  parsePixelSpacingMm,
  scaleSpacingToBitmap,
  screenToImage,
  segmentImageToScreen,
  spacingUsableForBitmap,
  type PacsMatrixSize,
  type PacsSegment,
  type PacsSpacingMm,
} from '~/utils/pacs-measure'

const props = defineProps<{
  leftInstanceId?: string
  rightInstanceId?: string
  compare?: boolean
}>()

type Tool = 'pan' | 'zoom' | 'wl' | 'measure' | 'arrow'

const { t } = useI18n()
const rootEl = ref<HTMLElement | null>(null)
const panesEl = ref<HTMLElement | null>(null)
const tool = ref<Tool>('pan')
const fullscreen = ref(false)
const frameIndex = ref(0)
const maxFrame = ref<number | null>(null)
const downloadBusy = ref(false)
/** Spacing from Orthanc metadata (DICOM pixels). */
const metaSpacing = ref<PacsSpacingMm | null>(null)
const metaMatrix = ref<PacsMatrixSize | null>(null)
/** Spacing trusted for the currently loaded left preview bitmap (null → px). */
const effectiveSpacing = ref<PacsSpacingMm | null>(null)
/** True when mm come from scaled spacing (preview ≠ DICOM matrix). */
const previewResized = ref(false)

const measureCalibrated = computed(() => effectiveSpacing.value != null)
const measureHint = computed(() => {
  if (tool.value !== 'measure') return ''
  if (measureCalibrated.value && previewResized.value) return t('pacs.tools.measureHintScaled')
  if (measureCalibrated.value) return ''
  return t('pacs.tools.measureHintPx')
})
const wlBadge = computed(() => `B ${Math.round(left.wl)} · C ${Math.round(left.ww)}`)

function refreshEffectiveSpacing(img: HTMLImageElement | null) {
  if (!img || !img.naturalWidth) {
    effectiveSpacing.value = null
    previewResized.value = false
    return
  }
  const exact = spacingUsableForBitmap(
    metaSpacing.value,
    metaMatrix.value,
    img.naturalWidth,
    img.naturalHeight,
  )
  if (exact) {
    effectiveSpacing.value = exact
    previewResized.value = false
    return
  }
  const scaled = scaleSpacingToBitmap(
    metaSpacing.value,
    metaMatrix.value,
    img.naturalWidth,
    img.naturalHeight,
  )
  effectiveSpacing.value = scaled
  previewResized.value = Boolean(scaled)
}

type Pane = {
  img: HTMLImageElement | null
  scale: number
  offsetX: number
  offsetY: number
  wl: number
  ww: number
  dragging: boolean
  lastX: number
  lastY: number
  /** Image-space segments (bitmap px, origin = image centre). */
  measure: PacsSegment | null
  arrow: PacsSegment | null
  drawing: boolean
}

function emptyPane(): Pane {
  return {
    img: null, scale: 1, offsetX: 0, offsetY: 0, wl: 0, ww: 100,
    dragging: false, lastX: 0, lastY: 0, measure: null, arrow: null, drawing: false,
  }
}

const left = reactive(emptyPane())
const right = reactive(emptyPane())
const leftCanvas = ref<HTMLCanvasElement | null>(null)
const rightCanvas = ref<HTMLCanvasElement | null>(null)

const imageCache = new Map<string, string>()
const loadError = ref('')
/** Bumps on each bind to ignore stale async preview responses. */
let loadGen = 0
let resizeObserver: ResizeObserver | null = null

function reportLoadError(message: string) {
  loadError.value = message
}

function fetchStatus(e: unknown): number | null {
  const err = e as { statusCode?: number, status?: number, response?: { status?: number } }
  const n = err?.statusCode ?? err?.status ?? err?.response?.status
  return typeof n === 'number' ? n : null
}

function clampWl(v: number) {
  return Math.max(-80, Math.min(80, v))
}
function clampWw(v: number) {
  return Math.max(20, Math.min(300, v))
}

function bumpContrast(deltaWl: number, deltaWw: number) {
  left.wl = clampWl(left.wl + deltaWl)
  left.ww = clampWw(left.ww + deltaWw)
  if (props.compare) {
    right.wl = left.wl
    right.ww = left.ww
  }
  paint(left, leftCanvas.value)
  if (props.compare) paint(right, rightCanvas.value)
}

function seedWlFromMetadata(windowCenter: unknown, windowWidth: unknown) {
  const wc = Number(windowCenter)
  const ww = Number(windowWidth)
  if (!Number.isFinite(wc) || !Number.isFinite(ww) || ww <= 0) return
  // Map DICOM W/L into canvas filter ranges (heuristic, preview already windowed).
  left.wl = clampWl(wc / 40)
  left.ww = clampWw(ww / 4)
  right.wl = left.wl
  right.ww = left.ww
}

async function loadPreview(instanceId: string, frame: number): Promise<string> {
  const key = `${instanceId}:${frame}`
  const cached = imageCache.get(key)
  if (cached) return cached
  const blob = await $fetch<Blob>(`/api/pacs/instances/${instanceId}/frames/${frame}/preview`, {
    responseType: 'blob',
  })
  const ct = (blob.type || '').toLowerCase()
  if (ct.includes('json') || ct.includes('text')) {
    throw new Error('preview_not_image')
  }
  const head = new Uint8Array(await blob.slice(0, 8).arrayBuffer())
  const isPng = head.length >= 8
    && head[0] === 0x89 && head[1] === 0x50 && head[2] === 0x4e && head[3] === 0x47
  const isJpeg = head.length >= 3 && head[0] === 0xff && head[1] === 0xd8 && head[2] === 0xff
  if (!isPng && !isJpeg && !ct.startsWith('image/')) {
    throw new Error('preview_not_image')
  }
  const url = URL.createObjectURL(blob)
  imageCache.set(key, url)
  return url
}

function looksLikeDicom(buf: Uint8Array): boolean {
  return buf.length >= 132
    && buf[128] === 0x44 && buf[129] === 0x49 && buf[130] === 0x43 && buf[131] === 0x4d
}

/** @returns true if a frame was painted successfully */
async function bindPane(
  pane: Pane,
  canvas: HTMLCanvasElement | null,
  instanceId?: string,
  gen?: number,
): Promise<boolean> {
  const myGen = gen ?? ++loadGen
  if (!canvas || !instanceId) {
    pane.img = null
    paint(pane, canvas)
    return false
  }
  const frame = frameIndex.value
  try {
    const url = await loadPreview(instanceId, frame)
    if (myGen !== loadGen) return false
    const img = new Image()
    await new Promise<void>((resolve, reject) => {
      img.onload = () => resolve()
      img.onerror = () => reject(new Error('preview_decode_failed'))
      img.src = url
    })
    if (myGen !== loadGen) return false
    pane.img = img
    pane.scale = 1
    pane.offsetX = 0
    pane.offsetY = 0
    if (pane === left) refreshEffectiveSpacing(img)
    loadError.value = ''
    paint(pane, canvas)
    return true
  } catch (e) {
    if (myGen !== loadGen) return false
    pane.img = null
    if (pane === left) refreshEffectiveSpacing(null)
    paint(pane, canvas)
    if (frame > 0 && fetchStatus(e) === 404) {
      maxFrame.value = frame - 1
      frameIndex.value = maxFrame.value
      return bindPane(pane, canvas, instanceId, myGen)
    }
    reportLoadError(fetchErrorMessage(e) || t('pacs.previewError'))
    return false
  }
}

function fetchErrorMessage(e: unknown): string {
  const err = e as { data?: any, message?: string }
  const d = err?.data
  const msg = d?.message || d?.error?.message || d?.statusMessage
  if (typeof msg === 'string' && msg && msg !== 'Error' && !msg.startsWith('errors.')) return msg
  const key = d?.msgKey || d?.error?.msgKey
  if (key === 'pacs_instance_unavailable') return t('pacs.instanceUnavailable')
  return ''
}

function paint(pane: Pane, canvas: HTMLCanvasElement | null) {
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const w = canvas.width
  const h = canvas.height
  ctx.clearRect(0, 0, w, h)
  ctx.fillStyle = '#0b1220'
  ctx.fillRect(0, 0, w, h)
  if (!pane.img) return
  const contrast = Math.max(0.2, pane.ww / 100)
  const brightness = 1 + pane.wl / 100
  ctx.save()
  ctx.filter = `brightness(${brightness}) contrast(${contrast})`
  ctx.translate(w / 2 + pane.offsetX, h / 2 + pane.offsetY)
  ctx.scale(pane.scale, pane.scale)
  ctx.drawImage(pane.img, -pane.img.width / 2, -pane.img.height / 2)
  ctx.restore()

  const drawLine = (seg: PacsSegment | null, color: string, arrow = false) => {
    if (!seg) return
    const line = segmentImageToScreen(seg, w, h, pane.offsetX, pane.offsetY, pane.scale)
    ctx.strokeStyle = color
    ctx.lineWidth = 2
    ctx.beginPath()
    ctx.moveTo(line.x1, line.y1)
    ctx.lineTo(line.x2, line.y2)
    ctx.stroke()
    if (arrow) {
      const ang = Math.atan2(line.y2 - line.y1, line.x2 - line.x1)
      ctx.beginPath()
      ctx.moveTo(line.x2, line.y2)
      ctx.lineTo(line.x2 - 10 * Math.cos(ang - 0.4), line.y2 - 10 * Math.sin(ang - 0.4))
      ctx.lineTo(line.x2 - 10 * Math.cos(ang + 0.4), line.y2 - 10 * Math.sin(ang + 0.4))
      ctx.closePath()
      ctx.fillStyle = color
      ctx.fill()
    } else {
      const label = formatPacsLength(measureImageSegment(
        seg.x2 - seg.x1,
        seg.y2 - seg.y1,
        effectiveSpacing.value,
      ))
      ctx.fillStyle = color
      ctx.font = '12px sans-serif'
      ctx.fillText(label, (line.x1 + line.x2) / 2 + 6, (line.y1 + line.y2) / 2 - 6)
    }
  }
  drawLine(pane.measure, '#5eead4')
  drawLine(pane.arrow, '#fbbf24', true)
}

function pointerToImage(pane: Pane, canvas: HTMLCanvasElement, e: PointerEvent) {
  return screenToImage(
    e.offsetX,
    e.offsetY,
    canvas.width,
    canvas.height,
    pane.offsetX,
    pane.offsetY,
    pane.scale,
  )
}

function onPointerDown(pane: Pane, canvas: HTMLCanvasElement | null, e: PointerEvent) {
  if (!canvas) return
  canvas.setPointerCapture(e.pointerId)
  pane.dragging = true
  pane.lastX = e.offsetX
  pane.lastY = e.offsetY
  if (tool.value === 'measure' || tool.value === 'arrow') {
    pane.drawing = true
    const p = pointerToImage(pane, canvas, e)
    const line = { x1: p.x, y1: p.y, x2: p.x, y2: p.y }
    if (tool.value === 'measure') pane.measure = line
    else pane.arrow = line
  }
}

function onPointerMove(pane: Pane, canvas: HTMLCanvasElement | null, e: PointerEvent) {
  if (!pane.dragging || !canvas) return
  const dx = e.offsetX - pane.lastX
  const dy = e.offsetY - pane.lastY
  if (tool.value === 'pan') {
    pane.offsetX += dx
    pane.offsetY += dy
  } else if (tool.value === 'zoom') {
    pane.scale = Math.min(8, Math.max(0.2, pane.scale * (1 + dy * -0.01)))
  } else if (tool.value === 'wl') {
    pane.wl = clampWl(pane.wl + dx * 0.3)
    pane.ww = clampWw(pane.ww + dy * -0.5)
  } else if (pane.drawing) {
    const line = tool.value === 'measure' ? pane.measure : pane.arrow
    if (line) {
      const p = pointerToImage(pane, canvas, e)
      line.x2 = p.x
      line.y2 = p.y
    }
  }
  pane.lastX = e.offsetX
  pane.lastY = e.offsetY
  paint(pane, canvas)
}

function onPointerUp(pane: Pane) {
  pane.dragging = false
  pane.drawing = false
}

async function toggleFullscreen() {
  if (!rootEl.value) return
  if (!document.fullscreenElement) {
    await rootEl.value.requestFullscreen()
  } else {
    await document.exitFullscreen()
  }
}

function syncFullscreenFlag() {
  fullscreen.value = document.fullscreenElement === rootEl.value
  nextTick(resizeCanvases)
}

function resizeCanvases() {
  for (const [el, pane] of [[leftCanvas.value, left], [rightCanvas.value, right]] as const) {
    if (!el) continue
    const parent = el.parentElement
    if (!parent) continue
    // Keep buffer size == CSS layout size so offsetX/Y map 1:1 (fullscreen/resize).
    el.width = Math.max(1, parent.clientWidth)
    el.height = Math.max(1, parent.clientHeight)
    paint(pane as Pane, el)
  }
}

async function reloadPanes(): Promise<boolean> {
  const gen = ++loadGen
  const leftOk = await bindPane(left, leftCanvas.value, props.leftInstanceId, gen)
  if (myGenStale(gen)) return false
  if (!props.compare) return leftOk
  const rightOk = await bindPane(right, rightCanvas.value, props.rightInstanceId, gen)
  return leftOk && rightOk
}

function myGenStale(gen: number): boolean {
  return gen !== loadGen
}

async function stepFrame(delta: number) {
  const prev = frameIndex.value
  const next = prev + delta
  if (next < 0) return
  if (maxFrame.value != null && next > maxFrame.value) return
  frameIndex.value = next
  const ok = await reloadPanes()
  if (!ok && frameIndex.value === next) {
    frameIndex.value = prev
    await reloadPanes()
  }
}

function onWheel(e: WheelEvent) {
  if (!e.shiftKey) return
  e.preventDefault()
  void stepFrame(e.deltaY > 0 ? 1 : -1)
}

async function downloadDicom() {
  const id = props.leftInstanceId
  if (!id || downloadBusy.value) return
  downloadBusy.value = true
  try {
    const blob = await $fetch<Blob>(`/api/pacs/instances/${id}/file`, { responseType: 'blob' })
    const head = new Uint8Array(await blob.slice(0, 132).arrayBuffer())
    if (!looksLikeDicom(head)) {
      throw new Error('not_dicom')
    }
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${id.slice(0, 12)}.dcm`
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    reportLoadError(t('pacs.downloadError'))
  } finally {
    downloadBusy.value = false
  }
}

watch(() => props.leftInstanceId, async (id) => {
  frameIndex.value = 0
  maxFrame.value = null
  metaSpacing.value = null
  metaMatrix.value = null
  effectiveSpacing.value = null
  previewResized.value = false
  left.measure = null
  left.arrow = null
  if (id) {
    try {
      const res: any = await $fetch(`/api/pacs/instances/${id}/metadata`)
      const data = res.data ?? res
      const n = Number(data.numberOfFrames)
      if (Number.isFinite(n) && n > 0) maxFrame.value = n - 1
      metaSpacing.value = parsePixelSpacingMm(data.pixelSpacingMm)
      metaMatrix.value = parseMatrixSize(data.rows, data.columns)
      seedWlFromMetadata(data.windowCenter, data.windowWidth)
    } catch { /* optional */ }
  }
  void bindPane(left, leftCanvas.value, id)
})
watch(() => props.rightInstanceId, (id) => {
  right.measure = null
  right.arrow = null
  void bindPane(right, rightCanvas.value, id)
})
watch(() => props.compare, () => nextTick(resizeCanvases))

onMounted(async () => {
  await nextTick()
  resizeCanvases()
  window.addEventListener('resize', resizeCanvases)
  document.addEventListener('fullscreenchange', syncFullscreenFlag)
  if (typeof ResizeObserver !== 'undefined' && panesEl.value) {
    resizeObserver = new ResizeObserver(() => resizeCanvases())
    resizeObserver.observe(panesEl.value)
  }
  await reloadPanes()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCanvases)
  document.removeEventListener('fullscreenchange', syncFullscreenFlag)
  resizeObserver?.disconnect()
  resizeObserver = null
  for (const url of imageCache.values()) URL.revokeObjectURL(url)
  imageCache.clear()
})
</script>

<template>
  <div ref="rootEl" class="dicom-viewer" data-testid="dicom-viewer">
    <p
      v-if="loadError"
      class="pro-inline-feedback pro-inline-feedback--error"
      role="alert"
      data-testid="dicom-preview-error"
    >
      {{ loadError }}
    </p>
    <div class="dicom-viewer__toolbar" role="toolbar">
      <button
        v-for="tb in (['pan','zoom','wl','measure','arrow'] as Tool[])"
        :key="tb"
        type="button"
        class="dicom-viewer__tool"
        :class="{ 'is-active': tool === tb }"
        :data-testid="`dicom-tool-${tb}`"
        @click="tool = tb"
      >
        {{ t(`pacs.tools.${tb}`) }}
      </button>
      <button
        type="button"
        class="dicom-viewer__tool"
        data-testid="dicom-contrast-down"
        :title="t('pacs.tools.contrastDown')"
        @click="bumpContrast(-8, -15)"
      >
        −
      </button>
      <button
        type="button"
        class="dicom-viewer__tool"
        data-testid="dicom-contrast-up"
        :title="t('pacs.tools.contrastUp')"
        @click="bumpContrast(8, 15)"
      >
        +
      </button>
      <span class="dicom-viewer__wl" data-testid="dicom-wl-badge">{{ wlBadge }}</span>
      <button
        type="button"
        class="dicom-viewer__tool"
        data-testid="dicom-tool-frame-prev"
        :disabled="frameIndex <= 0"
        @click="stepFrame(-1)"
      >
        {{ t('pacs.tools.prevFrame') }}
      </button>
      <span class="dicom-viewer__frame" data-testid="dicom-frame-label">
        {{ t('pacs.tools.frame', { n: frameIndex + 1 }) }}
      </span>
      <button
        type="button"
        class="dicom-viewer__tool"
        data-testid="dicom-tool-frame-next"
        :disabled="maxFrame != null && frameIndex >= maxFrame"
        @click="stepFrame(1)"
      >
        {{ t('pacs.tools.nextFrame') }}
      </button>
      <button
        type="button"
        class="dicom-viewer__tool"
        data-testid="dicom-tool-download"
        :disabled="!leftInstanceId || downloadBusy"
        @click="downloadDicom"
      >
        {{ t('pacs.tools.download') }}
      </button>
      <button
        type="button"
        class="dicom-viewer__tool"
        data-testid="dicom-tool-fullscreen"
        @click="toggleFullscreen"
      >
        {{ fullscreen ? t('pacs.tools.exitFullscreen') : t('pacs.tools.fullscreen') }}
      </button>
      <span
        class="dicom-viewer__calib"
        :data-calibrated="measureCalibrated ? '1' : '0'"
        data-testid="dicom-calib-badge"
      >
        {{ measureCalibrated ? t('pacs.tools.measureCalibrated') : t('pacs.tools.measureUncalibrated') }}
      </span>
    </div>
    <p
      v-if="measureHint"
      class="dicom-viewer__hint"
      data-testid="dicom-measure-hint"
    >
      {{ measureHint }}
    </p>
    <div
      ref="panesEl"
      class="dicom-viewer__panes"
      :class="{ 'dicom-viewer__panes--compare': compare }"
      @wheel="onWheel"
    >
      <div class="dicom-viewer__pane">
        <canvas
          ref="leftCanvas"
          data-testid="dicom-canvas-left"
          @pointerdown="onPointerDown(left, leftCanvas, $event)"
          @pointermove="onPointerMove(left, leftCanvas, $event)"
          @pointerup="onPointerUp(left)"
          @pointerleave="onPointerUp(left)"
        />
      </div>
      <div v-if="compare" class="dicom-viewer__pane">
        <canvas
          ref="rightCanvas"
          data-testid="dicom-canvas-right"
          @pointerdown="onPointerDown(right, rightCanvas, $event)"
          @pointermove="onPointerMove(right, rightCanvas, $event)"
          @pointerup="onPointerUp(right)"
          @pointerleave="onPointerUp(right)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.dicom-viewer {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 360px;
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  padding: 0.75rem;
}
.dicom-viewer:fullscreen {
  width: 100%;
  height: 100%;
  border-radius: 0;
  box-sizing: border-box;
}
.dicom-viewer:fullscreen .dicom-viewer__panes {
  flex: 1;
  min-height: 0;
}
.dicom-viewer:fullscreen .dicom-viewer__pane {
  min-height: 0;
  height: 100%;
}
.dicom-viewer__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.dicom-viewer__tool {
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-bg);
  color: var(--pf-vet-primary);
  border-radius: 8px;
  padding: 0.35rem 0.7rem;
  font-size: 0.85rem;
  cursor: pointer;
}
.dicom-viewer__tool:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.dicom-viewer__tool.is-active {
  background: var(--pf-vet-accent);
  color: #fff;
  border-color: transparent;
}
.dicom-viewer__frame,
.dicom-viewer__wl {
  font-size: 0.8rem;
  color: var(--pf-vet-primary);
  opacity: 0.85;
  min-width: 4.5rem;
  text-align: center;
}
.dicom-viewer__wl {
  font-variant-numeric: tabular-nums;
  border: 1px solid var(--pf-vet-border);
  border-radius: 6px;
  padding: 0.2rem 0.45rem;
}
.dicom-viewer__calib {
  font-size: 0.78rem;
  opacity: 0.85;
  color: var(--pf-vet-primary);
  border: 1px dashed var(--pf-vet-border);
  border-radius: 6px;
  padding: 0.2rem 0.45rem;
}
.dicom-viewer__calib[data-calibrated='1'] {
  border-color: var(--pf-vet-accent);
  color: var(--pf-vet-accent);
}
.dicom-viewer__hint {
  margin: 0;
  font-size: 0.8rem;
  opacity: 0.75;
  color: var(--pf-vet-primary);
}
.dicom-viewer__panes {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.5rem;
  flex: 1;
  min-height: 320px;
}
.dicom-viewer__panes--compare {
  grid-template-columns: 1fr 1fr;
}
.dicom-viewer__pane {
  position: relative;
  min-height: 320px;
  background: #0b1220;
  border-radius: 8px;
  overflow: hidden;
}
.dicom-viewer__pane canvas {
  width: 100%;
  height: 100%;
  display: block;
  touch-action: none;
  cursor: crosshair;
}
@media (max-width: 900px) {
  .dicom-viewer__panes--compare {
    grid-template-columns: 1fr;
  }
}
</style>
