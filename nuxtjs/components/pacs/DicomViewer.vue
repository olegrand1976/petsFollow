<script setup lang="ts">
/**
 * Canvas DICOM viewer (client-only).
 * Loads Orthanc frame previews via BFF; tools: zoom, pan, W/L, measure, arrow, fullscreen, dual-pane.
 * Cornerstone3D-ready structure — pixel source is Orthanc preview for V1 reliability (no WASM CDN).
 */
const props = defineProps<{
  leftInstanceId?: string
  rightInstanceId?: string
  compare?: boolean
}>()

type Tool = 'pan' | 'zoom' | 'wl' | 'measure' | 'arrow'

const { t } = useI18n()
const rootEl = ref<HTMLElement | null>(null)
const tool = ref<Tool>('pan')
const fullscreen = ref(false)

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
  measure: { x1: number, y1: number, x2: number, y2: number } | null
  arrow: { x1: number, y1: number, x2: number, y2: number } | null
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

async function loadPreview(instanceId: string): Promise<string> {
  const cached = imageCache.get(instanceId)
  if (cached) return cached
  const blob = await $fetch<Blob>(`/api/pacs/instances/${instanceId}/frames/0/preview`, {
    responseType: 'blob',
  })
  const url = URL.createObjectURL(blob)
  imageCache.set(instanceId, url)
  return url
}

async function bindPane(pane: Pane, canvas: HTMLCanvasElement | null, instanceId?: string) {
  if (!canvas || !instanceId) {
    pane.img = null
    paint(pane, canvas)
    return
  }
  const url = await loadPreview(instanceId)
  const img = new Image()
  img.onload = () => {
    pane.img = img
    pane.scale = 1
    pane.offsetX = 0
    pane.offsetY = 0
    paint(pane, canvas)
  }
  img.src = url
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

  const drawLine = (line: { x1: number, y1: number, x2: number, y2: number } | null, color: string, arrow = false) => {
    if (!line) return
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
      const dx = line.x2 - line.x1
      const dy = line.y2 - line.y1
      const px = Math.round(Math.hypot(dx, dy))
      ctx.fillStyle = color
      ctx.font = '12px sans-serif'
      ctx.fillText(`${px} px`, (line.x1 + line.x2) / 2 + 6, (line.y1 + line.y2) / 2 - 6)
    }
  }
  drawLine(pane.measure, '#5eead4')
  drawLine(pane.arrow, '#fbbf24', true)
}

function onPointerDown(pane: Pane, canvas: HTMLCanvasElement | null, e: PointerEvent) {
  if (!canvas) return
  canvas.setPointerCapture(e.pointerId)
  pane.dragging = true
  pane.lastX = e.offsetX
  pane.lastY = e.offsetY
  if (tool.value === 'measure' || tool.value === 'arrow') {
    pane.drawing = true
    const line = { x1: e.offsetX, y1: e.offsetY, x2: e.offsetX, y2: e.offsetY }
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
    pane.wl = Math.max(-80, Math.min(80, pane.wl + dx * 0.3))
    pane.ww = Math.max(20, Math.min(300, pane.ww + dy * -0.5))
  } else if (pane.drawing) {
    const line = tool.value === 'measure' ? pane.measure : pane.arrow
    if (line) {
      line.x2 = e.offsetX
      line.y2 = e.offsetY
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
    fullscreen.value = true
  } else {
    await document.exitFullscreen()
    fullscreen.value = false
  }
}

function resizeCanvases() {
  for (const [el, pane] of [[leftCanvas.value, left], [rightCanvas.value, right]] as const) {
    if (!el) continue
    const parent = el.parentElement
    if (!parent) continue
    el.width = parent.clientWidth
    el.height = Math.max(320, parent.clientHeight)
    paint(pane as Pane, el)
  }
}

watch(() => props.leftInstanceId, (id) => { void bindPane(left, leftCanvas.value, id) })
watch(() => props.rightInstanceId, (id) => { void bindPane(right, rightCanvas.value, id) })
watch(() => props.compare, () => nextTick(resizeCanvases))

onMounted(async () => {
  await nextTick()
  resizeCanvases()
  window.addEventListener('resize', resizeCanvases)
  await bindPane(left, leftCanvas.value, props.leftInstanceId)
  if (props.compare) await bindPane(right, rightCanvas.value, props.rightInstanceId)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCanvases)
  for (const url of imageCache.values()) URL.revokeObjectURL(url)
  imageCache.clear()
})
</script>

<template>
  <div ref="rootEl" class="dicom-viewer" data-testid="dicom-viewer">
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
        data-testid="dicom-tool-fullscreen"
        @click="toggleFullscreen"
      >
        {{ fullscreen ? t('pacs.tools.exitFullscreen') : t('pacs.tools.fullscreen') }}
      </button>
    </div>
    <div class="dicom-viewer__panes" :class="{ 'dicom-viewer__panes--compare': compare }">
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
.dicom-viewer__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
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
.dicom-viewer__tool.is-active {
  background: var(--pf-vet-accent);
  color: #fff;
  border-color: transparent;
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
