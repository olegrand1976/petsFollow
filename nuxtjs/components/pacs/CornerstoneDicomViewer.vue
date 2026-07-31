<script setup lang="ts">
/**
 * Cornerstone3D stack viewer (P2.1) — loads DICOM via BFF /file (same-origin cookies).
 * Opt-in via NUXT_PUBLIC_PACS_VIEWER_ENGINE=cornerstone ; canvas remains default.
 */
const props = defineProps<{
  leftInstanceId?: string
  rightInstanceId?: string
  compare?: boolean
}>()

const { t } = useI18n()
const leftEl = ref<HTMLDivElement | null>(null)
const rightEl = ref<HTMLDivElement | null>(null)
const loadError = ref('')
const busy = ref(false)
const metaLabel = ref('')

const ENGINE_ID = 'pf-pacs-cs-engine'
let renderingEngine: any = null
let initPromise: Promise<void> | null = null

async function ensureCornerstone() {
  if (initPromise) return initPromise
  initPromise = (async () => {
    const core = await import('@cornerstonejs/core')
    const tools = await import('@cornerstonejs/tools')
    const dicomImageLoader = await import('@cornerstonejs/dicom-image-loader')
    await core.init()
    await tools.init()
    dicomImageLoader.init({ maxWebWorkers: 1 })
  })()
  return initPromise
}

function imageIdFor(instanceId: string) {
  // Same-origin BFF — httpOnly cookies sent with wadouri XHR (credentials).
  const path = `/api/pacs/instances/${encodeURIComponent(instanceId)}/file`
  return `wadouri:${window.location.origin}${path}`
}

async function loadMetadata(instanceId: string) {
  try {
    const res: any = await $fetch(`/api/pacs/instances/${instanceId}/metadata`)
    const data = res.data ?? res
    const spacing = Array.isArray(data.pixelSpacingMm) ? data.pixelSpacingMm.join('×') : null
    const parts = [
      data.modality,
      spacing ? `${spacing} mm` : null,
      data.numberOfFrames ? `${data.numberOfFrames} frames` : null,
    ].filter(Boolean)
    metaLabel.value = parts.join(' · ')
  } catch {
    metaLabel.value = ''
  }
}

async function bindStack(element: HTMLDivElement | null, instanceId: string | undefined, viewportId: string) {
  if (!element || !instanceId || !renderingEngine) return
  const core = await import('@cornerstonejs/core')
  const { Enums } = core
  element.innerHTML = ''
  renderingEngine.enableElement({
    viewportId,
    element,
    type: Enums.ViewportType.STACK,
  })
  const vp = renderingEngine.getViewport(viewportId)
  await vp.setStack([imageIdFor(instanceId)])
  vp.render()
}

async function reload() {
  if (!props.leftInstanceId) return
  busy.value = true
  loadError.value = ''
  try {
    await ensureCornerstone()
    const core = await import('@cornerstonejs/core')
    if (renderingEngine) {
      try { renderingEngine.destroy() } catch { /* ignore */ }
      renderingEngine = null
    }
    renderingEngine = new core.RenderingEngine(ENGINE_ID)
    await loadMetadata(props.leftInstanceId)
    await bindStack(leftEl.value, props.leftInstanceId, 'left')
    if (props.compare && props.rightInstanceId) {
      await bindStack(rightEl.value, props.rightInstanceId, 'right')
    }
  } catch (e: any) {
    loadError.value = e?.message || t('pacs.previewError')
  } finally {
    busy.value = false
  }
}

watch(
  () => [props.leftInstanceId, props.rightInstanceId, props.compare] as const,
  () => { void nextTick(() => void reload()) },
)

onMounted(() => { void reload() })

onBeforeUnmount(() => {
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
    <div class="cs-viewer__toolbar">
      <span class="cs-viewer__badge" data-testid="dicom-engine-badge">Cornerstone3D</span>
      <span v-if="metaLabel" class="cs-viewer__meta" data-testid="dicom-meta-label">{{ metaLabel }}</span>
      <span v-if="busy" class="cs-viewer__busy">{{ t('pacs.launch.waking') }}</span>
    </div>
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
  gap: 0.75rem;
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
.cs-viewer__meta { opacity: 0.75; color: var(--pf-vet-primary); }
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
}
@media (max-width: 900px) {
  .cs-viewer__panes--compare { grid-template-columns: 1fr; }
}
</style>
