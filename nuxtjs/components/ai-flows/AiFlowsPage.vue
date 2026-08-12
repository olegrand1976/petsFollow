<template>
  <div class="pf-ai-flows" data-testid="ai-flows-page">
    <ProPageHeader
      :title="$t('aiFlows.title')"
      :subtitle="$t('aiFlows.subtitle')"
    />

    <ProCard class="pf-ai-flows__howto">
      <h2 class="pf-ai-flows__h2">{{ $t('aiFlows.howtoTitle') }}</h2>
      <ol class="pf-ai-flows__howto-list">
        <li>{{ $t('aiFlows.howto1') }}</li>
        <li>{{ $t('aiFlows.howto2') }}</li>
        <li>{{ $t('aiFlows.howto3') }}</li>
      </ol>
    </ProCard>

    <h2 id="ai-flows-profiles-label" class="pf-ai-flows__h2">{{ $t('aiFlows.selectProfile') }}</h2>
    <div
      class="pf-ai-flows__profiles"
      data-testid="ai-flows-profiles"
      role="radiogroup"
      aria-labelledby="ai-flows-profiles-label"
      @keydown="onProfileKeydown"
    >
      <button
        v-for="p in profiles"
        :key="p.id"
        type="button"
        class="pro-card pro-card--interactive pf-ai-flows__profile"
        :class="{ 'pf-ai-flows__profile--active': selectedId === p.id }"
        role="radio"
        :aria-checked="selectedId === p.id"
        :tabindex="selectedId === p.id ? 0 : -1"
        :data-testid="`ai-flow-profile-${p.id}`"
        @click="selectProfile(p.id)"
      >
        <ProIcon :name="p.icon" :size="28" />
        <strong>{{ $t(`aiFlows.profiles.${p.id}.label`) }}</strong>
        <span class="pf-ai-flows__profile-mission">{{ $t(`aiFlows.profiles.${p.id}.mission`) }}</span>
        <ProBadge v-if="p.id === 'vet'" variant="success">{{ $t('aiFlows.vetHighlight') }}</ProBadge>
      </button>
    </div>

    <div v-if="diagrams.length" class="pf-ai-flows__layout">
      <ul class="pf-ai-flows__list" data-testid="ai-flows-diagram-list">
        <li v-for="d in diagrams" :key="d.id">
          <button
            type="button"
            class="pf-ai-flows__list-btn"
            :class="{ 'pf-ai-flows__list-btn--active': d.id === activeDiagramId }"
            :data-testid="`ai-flow-diagram-${d.id}`"
            @click="activeDiagramId = d.id"
          >
            <strong>{{ $t(`aiFlows.diagrams.${d.id}.title`) }}</strong>
            <span>{{ $t(`aiFlows.diagrams.${d.id}.desc`) }}</span>
          </button>
        </li>
      </ul>

      <ProCard v-if="activeDiagram" class="pf-ai-flows__panel" data-testid="ai-flows-panel">
        <h2 class="pf-ai-flows__h2 pf-ai-flows__h2--flush">
          {{ $t(`aiFlows.diagrams.${activeDiagram.id}.title`) }}
        </h2>
        <p class="pf-ai-flows__blurb">{{ $t(`aiFlows.diagrams.${activeDiagram.id}.desc`) }}</p>
        <ClientOnly>
          <AiFlowsMermaidDiagram
            :key="`${activeDiagram.id}-${locale}`"
            :diagram-id="activeDiagram.id"
            :source="resolvedSource"
            :ariaLabel="$t(`aiFlows.diagrams.${activeDiagram.id}.title`)"
          />
          <template #fallback>
            <p class="pro-hint">{{ $t('aiFlows.loading') }}</p>
          </template>
        </ClientOnly>
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  aiFlowsCatalog,
  diagramsForProfile,
  fillMermaidTemplate,
  parseAiFlowProfileId,
} from '~/data/ai-flows/catalog'
import type { AiFlowProfileId } from '~/data/ai-flows/types'

const { tm, rt, locale } = useI18n()
const route = useRoute()
const router = useRouter()

const profiles = aiFlowsCatalog.profiles
const profileIds = profiles.map((p) => p.id)

const selectedId = computed<AiFlowProfileId>(
  () => parseAiFlowProfileId(route.query.profile) ?? 'vet',
)

const diagrams = computed(() => diagramsForProfile(selectedId.value))

const activeDiagramId = ref('')

watch(
  diagrams,
  (list) => {
    if (!list.length) {
      activeDiagramId.value = ''
      return
    }
    if (!list.some((d) => d.id === activeDiagramId.value)) {
      activeDiagramId.value = list[0]!.id
    }
  },
  { immediate: true },
)

const activeDiagram = computed(() =>
  diagrams.value.find((d) => d.id === activeDiagramId.value) ?? null,
)

function resolveMsg(value: unknown): string {
  if (typeof value === 'string') return value
  if (value == null) return ''
  try {
    return rt(value as never)
  } catch {
    return String(value)
  }
}

const resolvedSource = computed(() => {
  const d = activeDiagram.value
  if (!d) return ''
  // Depend on locale so labels refresh on language change.
  void locale.value
  const raw = tm(`aiFlows.diagrams.${d.id}.nodes`) as unknown
  const labels: Record<string, string> = {}
  if (raw && typeof raw === 'object' && !Array.isArray(raw)) {
    for (const [k, v] of Object.entries(raw as Record<string, unknown>)) {
      labels[k] = resolveMsg(v)
    }
  }
  return fillMermaidTemplate(d.diagram, labels)
})

function selectProfile(id: AiFlowProfileId) {
  if (selectedId.value === id && route.query.profile === id) return
  void router.replace({ path: '/flux-ia', query: { profile: id } })
}

function focusProfileButton(id: AiFlowProfileId) {
  void nextTick(() => {
    const el = document.querySelector<HTMLElement>(`[data-testid="ai-flow-profile-${id}"]`)
    el?.focus()
  })
}

function onProfileKeydown(e: KeyboardEvent) {
  const idx = profileIds.indexOf(selectedId.value)
  if (idx < 0) return
  let next: AiFlowProfileId | null = null
  switch (e.key) {
    case 'ArrowRight':
    case 'ArrowDown':
      next = profileIds[(idx + 1) % profileIds.length] ?? null
      break
    case 'ArrowLeft':
    case 'ArrowUp':
      next = profileIds[(idx - 1 + profileIds.length) % profileIds.length] ?? null
      break
    case 'Home':
      next = profileIds[0] ?? null
      break
    case 'End':
      next = profileIds[profileIds.length - 1] ?? null
      break
    default:
      return
  }
  if (!next) return
  e.preventDefault()
  selectProfile(next)
  focusProfileButton(next)
}
</script>

<style scoped>
.pf-ai-flows__howto {
  margin-bottom: 1.25rem;
}
.pf-ai-flows__h2 {
  margin: 0 0 0.75rem;
  font-size: 1.125rem;
  color: var(--pf-vet-primary);
}
.pf-ai-flows__h2--flush {
  margin-bottom: 0.35rem;
}
.pf-ai-flows__howto-list {
  margin: 0;
  padding-left: 1.25rem;
  line-height: 1.55;
}
.pf-ai-flows__profiles {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}
.pf-ai-flows__profile {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
  text-align: left;
  cursor: pointer;
  border: 1px solid var(--pf-vet-border);
  padding: 0.85rem;
}
.pf-ai-flows__profile--active {
  border-color: var(--pf-vet-accent);
  box-shadow: var(--pf-vet-shadow-sm);
}
.pf-ai-flows__profile-mission {
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
  line-height: 1.35;
}
.pf-ai-flows__layout {
  display: grid;
  grid-template-columns: minmax(14rem, 18rem) 1fr;
  gap: 1rem;
  align-items: start;
}
.pf-ai-flows__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.pf-ai-flows__list-btn {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  width: 100%;
  text-align: left;
  padding: 0.75rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-surface);
  cursor: pointer;
  color: inherit;
}
.pf-ai-flows__list-btn strong {
  color: var(--pf-vet-primary);
  font-size: 0.95rem;
}
.pf-ai-flows__list-btn span {
  font-size: 0.8rem;
  color: var(--pf-vet-muted, #64748b);
  line-height: 1.35;
}
.pf-ai-flows__list-btn--active {
  border-color: var(--pf-vet-accent);
  box-shadow: var(--pf-vet-shadow-sm);
}
.pf-ai-flows__blurb {
  margin: 0 0 1rem;
  color: var(--pf-vet-muted, #475569);
  line-height: 1.5;
}
@media (max-width: 860px) {
  .pf-ai-flows__layout {
    grid-template-columns: 1fr;
  }
}
</style>
