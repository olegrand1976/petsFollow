<template>
  <div class="pf-pres" data-testid="presentation-wizard">
    <ProPageHeader
      :title="$t('presentation.ui.title')"
      :subtitle="$t('presentation.ui.subtitle')"
    />

    <nav class="pf-pres__toc" :aria-label="$t('presentation.ui.tocLabel')" data-testid="presentation-toc">
      <button
        v-for="(id, idx) in steps"
        :key="id"
        type="button"
        class="pf-pres__toc-item"
        :class="{ 'pf-pres__toc-item--active': id === stepId }"
        :data-testid="`presentation-toc-${id}`"
        @click="go(id)"
      >
        <span class="pf-pres__toc-n">{{ idx + 1 }}</span>
        <span class="pf-pres__toc-label">{{ $t(`presentation.steps.${id}.nav`) }}</span>
      </button>
    </nav>

    <ProCard class="pf-pres__card" data-testid="presentation-step">
      <div class="pf-pres__meta">
        <span class="pf-pres__kicker">{{ $t(`presentation.steps.${stepId}.kicker`) }}</span>
        <span class="pf-pres__counter" data-testid="presentation-counter">
          {{ stepIndex + 1 }} / {{ steps.length }}
        </span>
      </div>
      <h2 class="pf-pres__headline">{{ $t(`presentation.steps.${stepId}.headline`) }}</h2>
      <p class="pf-pres__lead">{{ $t(`presentation.steps.${stepId}.lead`) }}</p>

      <ul v-if="points.length" class="pf-pres__points" data-testid="presentation-points">
        <li v-for="(p, i) in points" :key="i" class="pf-pres__point">
          <ProIcon :name="pointIcon(i)" :size="20" class="pf-pres__point-icon" />
          <div>
            <strong v-if="p.title">{{ p.title }}</strong>
            <span>{{ p.desc }}</span>
          </div>
        </li>
      </ul>

      <div v-if="stepId === 'offer'" class="pf-pres__offer" data-testid="presentation-offer">
        <div class="pf-pres__offer-grid">
          <div class="pf-pres__offer-card">
            <span class="pf-pres__offer-badge">{{ $t('products.pro.badge') }}</span>
            <strong>{{ $t('products.pro.price') }}</strong>
            <span>{{ $t('products.saas.longTerm.price') }}</span>
            <span class="pro-hint">{{ $t('products.saas.setup.price') }}</span>
          </div>
          <div class="pf-pres__offer-card">
            <span class="pf-pres__offer-badge">{{ $t('presentation.steps.offer.clientBadge') }}</span>
            <strong>{{ $t('products.plans.monthly.price') }} / {{ $t('products.plans.monthly.period') }}</strong>
            <span>{{ $t('products.plans.annual.price') }} · {{ $t('products.plans.triennial.price') }}</span>
            <span class="pro-hint">{{ $t('products.recommended') }} — {{ $t('products.plans.triennial.name') }}</span>
          </div>
        </div>
        <p class="pro-hint">{{ $t('products.partnerTip') }}</p>
      </div>

      <div v-if="stepId === 'ai_in_app' || stepId === 'ai_automation'" class="pf-pres__ai-cta">
        <ProButton test-id="presentation-open-ai-flows" variant="secondary" @click="navigateTo('/flux-ia')">
          {{ $t('presentation.ui.openAiFlows') }}
        </ProButton>
      </div>

      <div v-if="stepId === 'close'" class="pf-pres__close-cta">
        <ProButton test-id="presentation-cta" @click="onCloseCta">
          {{ $t('presentation.steps.close.cta') }}
        </ProButton>
        <ProButton test-id="presentation-cta-ai" variant="ghost" @click="navigateTo('/flux-ia')">
          {{ $t('presentation.ui.openAiFlows') }}
        </ProButton>
      </div>
    </ProCard>

    <footer class="pf-pres__footer">
      <ProButton
        variant="ghost"
        test-id="presentation-prev"
        :disabled="stepIndex === 0"
        @click="prev"
      >
        <ProIcon name="arrow_back" :size="16" />
        {{ $t('presentation.ui.prev') }}
      </ProButton>
      <ProButton
        test-id="presentation-next"
        :disabled="stepIndex >= steps.length - 1"
        @click="next"
      >
        {{ $t('presentation.ui.next') }}
        <ProIcon name="arrow_forward" :size="16" />
      </ProButton>
    </footer>
  </div>
</template>

<script setup lang="ts">
import {
  PRESENTATION_STEPS,
  parsePresentationStepId,
  type PresentationStepId,
} from '~/data/presentation/catalog'

const { t, tm, rt } = useI18n()
const route = useRoute()
const router = useRouter()
const { user } = useProUser()
const { isStagingLike } = useAppEnv()

const steps = PRESENTATION_STEPS

const stepId = computed<PresentationStepId>(
  () => parsePresentationStepId(route.query.step) ?? 'welcome',
)

const stepIndex = computed(() => steps.indexOf(stepId.value))

type Point = { title?: string; desc: string }

function resolveMsg(value: unknown): string {
  if (typeof value === 'string') return value
  if (value == null) return ''
  try {
    return rt(value as never)
  } catch {
    return String(value)
  }
}

const points = computed<Point[]>(() => {
  const raw = tm(`presentation.steps.${stepId.value}.points`) as unknown
  if (!Array.isArray(raw)) return []
  return raw.map((item) => {
    if (typeof item === 'string') return { desc: resolveMsg(item) }
    if (item && typeof item === 'object') {
      const o = item as Record<string, unknown>
      const title = o.title != null ? resolveMsg(o.title) : ''
      return {
        title: title || undefined,
        desc: resolveMsg(o.desc),
      }
    }
    return { desc: resolveMsg(item) }
  })
})

const POINT_ICONS = [
  'check_circle',
  'medical_services',
  'chat',
  'auto_awesome',
  'devices',
  'payments',
] as const

function pointIcon(i: number) {
  return POINT_ICONS[i % POINT_ICONS.length] ?? 'check_circle'
}

function go(id: PresentationStepId) {
  if (stepId.value === id && route.query.step === id) return
  void router.replace({ path: '/presentation', query: { step: id } })
}

function next() {
  const i = stepIndex.value
  if (i < steps.length - 1) go(steps[i + 1]!)
}

function prev() {
  const i = stepIndex.value
  if (i > 0) go(steps[i - 1]!)
}

function onCloseCta() {
  const role = user.value?.role
  if (role === 'commercial') {
    void navigateTo('/commercial/prospects')
    return
  }
  if (role === 'commercial_manager') {
    void navigateTo('/commercial-manager/prospects')
    return
  }
  if (isStagingLike.value) {
    void navigateTo('/usecases')
    return
  }
  void navigateTo('/admin')
}

useHead({
  title: computed(() => `${t(`presentation.steps.${stepId.value}.nav`)} — ${t('presentation.ui.title')}`),
})
</script>

<style scoped>
.pf-pres__toc {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 1.25rem;
}
.pf-pres__toc-item {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.65rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-surface);
  color: var(--pf-vet-primary);
  font-size: 0.8rem;
  cursor: pointer;
}
.pf-pres__toc-item--active {
  border-color: var(--pf-vet-accent);
  background: color-mix(in srgb, var(--pf-vet-accent) 12%, white);
  font-weight: 600;
}
.pf-pres__toc-n {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 999px;
  background: var(--pf-vet-primary);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
}
.pf-pres__toc-item--active .pf-pres__toc-n {
  background: var(--pf-vet-accent);
}
.pf-pres__card {
  margin-bottom: 1rem;
}
.pf-pres__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}
.pf-pres__kicker {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--pf-vet-accent);
}
.pf-pres__counter {
  font-variant-numeric: tabular-nums;
  font-size: 0.85rem;
  color: var(--pf-vet-muted, #64748b);
}
.pf-pres__headline {
  margin: 0 0 0.65rem;
  font-size: 1.5rem;
  color: var(--pf-vet-primary);
  line-height: 1.25;
}
.pf-pres__lead {
  margin: 0 0 1.25rem;
  color: var(--pf-vet-muted, #475569);
  line-height: 1.55;
  max-width: 42rem;
}
.pf-pres__points {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 0.75rem;
}
.pf-pres__point {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  padding: 0.75rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-bg, #f8fafc);
}
.pf-pres__point-icon {
  color: var(--pf-vet-accent);
  flex-shrink: 0;
  margin-top: 0.1rem;
}
.pf-pres__point strong {
  display: block;
  margin-bottom: 0.15rem;
  color: var(--pf-vet-primary);
}
.pf-pres__point span {
  font-size: 0.95rem;
  line-height: 1.45;
  color: var(--pf-vet-muted, #475569);
}
.pf-pres__offer {
  margin-top: 1rem;
}
.pf-pres__offer-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}
.pf-pres__offer-card {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 0.9rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-surface);
}
.pf-pres__offer-badge {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--pf-vet-accent);
}
.pf-pres__ai-cta,
.pf-pres__close-cta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-top: 1.25rem;
}
.pf-pres__footer {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  padding-top: 0.25rem;
}
@media (max-width: 640px) {
  .pf-pres__toc-label {
    display: none;
  }
}
</style>
