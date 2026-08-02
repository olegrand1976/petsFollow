<template>
  <div class="pf-mermaid" data-testid="mermaid-diagram">
    <div v-if="error" class="pf-mermaid__error" role="alert">{{ error }}</div>
    <div
      ref="hostEl"
      class="pf-mermaid__host"
      role="img"
      :aria-label="ariaLabel"
      data-testid="mermaid-svg-host"
    />
  </div>
</template>

<script setup lang="ts">
type MermaidClient = {
  initialize: (cfg: Record<string, unknown>) => void
  render: (id: string, text: string) => Promise<{ svg: string }>
}

const props = defineProps<{
  source: string
  diagramId: string
  ariaLabel: string
}>()

const { t } = useI18n()
const hostEl = ref<HTMLElement | null>(null)
const error = ref('')
let renderSeq = 0
let mermaidReady: Promise<MermaidClient> | null = null

async function getMermaid() {
  if (!mermaidReady) {
    mermaidReady = import('mermaid').then((mod) => {
      const mermaid = mod.default as MermaidClient
      mermaid.initialize({
        startOnLoad: false,
        securityLevel: 'strict',
        theme: 'neutral',
        fontFamily: 'inherit',
      })
      return mermaid
    })
  }
  return mermaidReady
}

async function render() {
  if (!import.meta.client || !hostEl.value) return
  const seq = ++renderSeq
  error.value = ''
  hostEl.value.innerHTML = ''
  try {
    const mermaid = await getMermaid()
    const id = `pf-mmd-${props.diagramId}-${seq}`
    const { svg } = await mermaid.render(id, props.source)
    if (seq !== renderSeq || !hostEl.value) return
    hostEl.value.innerHTML = svg
  } catch (e) {
    if (seq !== renderSeq) return
    error.value =
      e instanceof Error && e.message
        ? e.message
        : t('aiFlows.renderError')
  }
}

onMounted(() => {
  void render()
})

watch(
  () => [props.source, props.diagramId] as const,
  () => {
    void render()
  },
)
</script>

<style scoped>
.pf-mermaid {
  overflow-x: auto;
  padding: 0.75rem;
  background: var(--pf-vet-bg, #f8fafc);
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
}
.pf-mermaid__host :deep(svg) {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 0 auto;
}
.pf-mermaid__error {
  color: var(--pf-vet-alert, #c2410c);
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
}
</style>
