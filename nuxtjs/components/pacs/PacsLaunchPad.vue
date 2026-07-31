<script setup lang="ts">
/**
 * Startup sequence UI while Orthanc cold-starts (rocket-launch metaphor).
 * Steps follow real state (wake → starting → ready), with elapsed time only
 * to advance mid-flight steps — no fake percentage.
 */
const props = defineProps<{
  state: 'offline' | 'starting' | 'ready'
  waking: boolean
  /** Keep pad visible briefly after ready so the final step is seen. */
  celebrating?: boolean
}>()

const { t } = useI18n()

const STEPS = ['check', 'ignite', 'liftoff', 'orbit', 'ready'] as const

const startedAt = ref<number | null>(null)
const tick = ref(0)
let tickTimer: ReturnType<typeof setInterval> | null = null

const inFlight = computed(() =>
  props.waking || props.state === 'starting' || !!props.celebrating,
)

watch(
  inFlight,
  (on) => {
    if (on) {
      if (startedAt.value == null) startedAt.value = Date.now()
      if (!tickTimer) {
        tickTimer = setInterval(() => { tick.value += 1 }, 500)
      }
    } else {
      startedAt.value = null
      tick.value = 0
      if (tickTimer) {
        clearInterval(tickTimer)
        tickTimer = null
      }
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  if (tickTimer) clearInterval(tickTimer)
})

const elapsedMs = computed(() => {
  void tick.value
  if (startedAt.value == null) return 0
  return Date.now() - startedAt.value
})

/**
 * Discrete step index driven by Orthanc state:
 * - wake / offline after wake → ignite
 * - starting → liftoff then orbit (time-based within starting only)
 * - ready / celebrating → ready (all prior done)
 */
const currentIndex = computed(() => {
  if (props.state === 'ready' || props.celebrating) return STEPS.length - 1
  if (props.waking && props.state !== 'starting') return 1
  if (props.state === 'starting') {
    const e = elapsedMs.value
    if (e < 4000) return 2
    return 3
  }
  return 0
})

function stepStatus(i: number): 'done' | 'active' | 'pending' {
  if (props.state === 'ready' || props.celebrating) return 'done'
  if (i < currentIndex.value) return 'done'
  if (i === currentIndex.value) return 'active'
  return 'pending'
}

const rocketLifting = computed(() =>
  props.state === 'starting' || props.celebrating || (props.state === 'ready'),
)
</script>

<template>
  <div
    class="pacs-launch"
    data-testid="pacs-launch-pad"
    role="status"
    aria-live="polite"
    :aria-busy="inFlight && state !== 'ready'"
  >
    <div class="pacs-launch__visual" aria-hidden="true">
      <div class="pacs-launch__pad">
        <div
          class="pacs-launch__rocket"
          :class="{
            'is-igniting': waking && state !== 'starting' && !celebrating,
            'is-lifting': rocketLifting,
          }"
        >
          <svg viewBox="0 0 48 80" width="48" height="80" fill="none">
            <path
              d="M24 4 C18 18 16 34 16 48 L12 64 L24 58 L36 64 L32 48 C32 34 30 18 24 4Z"
              fill="currentColor"
              opacity="0.92"
            />
            <circle cx="24" cy="28" r="5" fill="var(--pf-vet-bg, #fff)" opacity="0.85" />
            <path d="M16 48 L8 56 L16 52Z" fill="currentColor" opacity="0.7" />
            <path d="M32 48 L40 56 L32 52Z" fill="currentColor" opacity="0.7" />
          </svg>
          <div class="pacs-launch__flame" />
        </div>
        <div class="pacs-launch__ground" />
      </div>
      <div
        class="pacs-launch__meter"
        :class="{ 'is-indeterminate': state === 'starting' || waking, 'is-complete': state === 'ready' || celebrating }"
      >
        <div class="pacs-launch__meter-fill" />
      </div>
      <p class="pacs-launch__phase">
        {{ t(`pacs.launch.steps.${STEPS[currentIndex]}`) }}
      </p>
    </div>

    <div class="pacs-launch__copy">
      <h4 class="pacs-launch__title">{{ t('pacs.launch.title') }}</h4>
      <p class="pacs-launch__subtitle">{{ t('pacs.launch.subtitle') }}</p>
      <ol class="pacs-launch__steps">
        <li
          v-for="(id, i) in STEPS"
          :key="id"
          class="pacs-launch__step"
          :class="`is-${stepStatus(i)}`"
          :data-testid="`pacs-launch-step-${id}`"
        >
          <span class="pacs-launch__bullet" aria-hidden="true" />
          <span>{{ t(`pacs.launch.steps.${id}`) }}</span>
        </li>
      </ol>
    </div>
  </div>
</template>

<style scoped>
.pacs-launch {
  display: grid;
  grid-template-columns: minmax(140px, 180px) 1fr;
  gap: 1.25rem;
  align-items: center;
  padding: 1.25rem 1.35rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 12px;
  background:
    radial-gradient(ellipse 80% 60% at 20% 100%, color-mix(in srgb, var(--pf-vet-accent) 18%, transparent), transparent 70%),
    var(--pf-vet-surface);
}
.pacs-launch__visual {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.65rem;
}
.pacs-launch__pad {
  position: relative;
  width: 100px;
  height: 110px;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}
.pacs-launch__rocket {
  position: relative;
  color: var(--pf-vet-primary);
  transform: translateY(8px);
  transition: transform 0.8s ease-out;
  z-index: 1;
}
.pacs-launch__rocket.is-igniting {
  animation: pacs-rocket-shake 0.35s ease-in-out infinite;
}
.pacs-launch__rocket.is-lifting {
  transform: translateY(-28px);
  animation: pacs-rocket-lift 2.4s ease-in-out infinite alternate;
}
.pacs-launch__flame {
  position: absolute;
  left: 50%;
  bottom: -2px;
  width: 14px;
  height: 0;
  margin-left: -7px;
  border-radius: 0 0 8px 8px;
  background: linear-gradient(180deg, #fbbf24, #ea580c 70%, transparent);
  opacity: 0;
  z-index: 0;
  pointer-events: none;
}
.pacs-launch__rocket.is-igniting .pacs-launch__flame,
.pacs-launch__rocket.is-lifting .pacs-launch__flame {
  height: 22px;
  opacity: 1;
  animation: pacs-flame 0.2s ease-in-out infinite alternate;
}
.pacs-launch__ground {
  position: absolute;
  bottom: 0;
  left: 10%;
  right: 10%;
  height: 4px;
  border-radius: 2px;
  background: color-mix(in srgb, var(--pf-vet-primary) 35%, transparent);
}
.pacs-launch__meter {
  width: 100%;
  height: 6px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--pf-vet-border) 80%, transparent);
  overflow: hidden;
}
.pacs-launch__meter-fill {
  height: 100%;
  width: 0;
  border-radius: inherit;
  background: var(--pf-vet-accent);
}
.pacs-launch__meter.is-indeterminate .pacs-launch__meter-fill {
  width: 40%;
  animation: pacs-indet 1.2s ease-in-out infinite;
}
.pacs-launch__meter.is-complete .pacs-launch__meter-fill {
  width: 100%;
  animation: none;
}
.pacs-launch__phase {
  margin: 0;
  font-size: 0.72rem;
  text-align: center;
  color: var(--pf-vet-primary);
  opacity: 0.75;
  line-height: 1.3;
  max-width: 11rem;
}
.pacs-launch__title {
  margin: 0 0 0.25rem;
  font-size: 1.05rem;
  color: var(--pf-vet-primary);
}
.pacs-launch__subtitle {
  margin: 0 0 0.85rem;
  font-size: 0.9rem;
  opacity: 0.8;
  color: var(--pf-vet-primary);
}
.pacs-launch__steps {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}
.pacs-launch__step {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  font-size: 0.88rem;
  color: var(--pf-vet-primary);
  opacity: 0.45;
  transition: opacity 0.25s ease;
}
.pacs-launch__step.is-done,
.pacs-launch__step.is-active {
  opacity: 1;
}
.pacs-launch__bullet {
  width: 0.65rem;
  height: 0.65rem;
  border-radius: 50%;
  border: 2px solid var(--pf-vet-border);
  flex-shrink: 0;
}
.pacs-launch__step.is-done .pacs-launch__bullet {
  background: var(--pf-vet-accent);
  border-color: var(--pf-vet-accent);
}
.pacs-launch__step.is-active .pacs-launch__bullet {
  border-color: var(--pf-vet-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--pf-vet-accent) 28%, transparent);
  animation: pacs-pulse 1s ease-in-out infinite;
}
@keyframes pacs-rocket-shake {
  from { transform: translateY(8px) translateX(-1px); }
  to { transform: translateY(8px) translateX(1px); }
}
@keyframes pacs-rocket-lift {
  from { transform: translateY(-22px); }
  to { transform: translateY(-36px); }
}
@keyframes pacs-flame {
  from { height: 16px; opacity: 0.75; }
  to { height: 26px; opacity: 1; }
}
@keyframes pacs-pulse {
  from { box-shadow: 0 0 0 2px color-mix(in srgb, var(--pf-vet-accent) 20%, transparent); }
  to { box-shadow: 0 0 0 5px color-mix(in srgb, var(--pf-vet-accent) 8%, transparent); }
}
@keyframes pacs-indet {
  0% { transform: translateX(-120%); }
  100% { transform: translateX(280%); }
}
@media (max-width: 640px) {
  .pacs-launch {
    grid-template-columns: 1fr;
    text-align: center;
  }
  .pacs-launch__steps { align-items: center; }
  .pacs-launch__step { justify-content: center; }
}
@media (prefers-reduced-motion: reduce) {
  .pacs-launch__rocket.is-igniting,
  .pacs-launch__rocket.is-lifting,
  .pacs-launch__flame,
  .pacs-launch__step.is-active .pacs-launch__bullet,
  .pacs-launch__meter.is-indeterminate .pacs-launch__meter-fill {
    animation: none;
  }
  .pacs-launch__meter.is-indeterminate .pacs-launch__meter-fill {
    width: 55%;
  }
}
</style>
