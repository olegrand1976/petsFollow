<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="pro-modal"
      :data-testid="testId"
      :style="{ zIndex: String(1000 + stackDepth) }"
      @keydown.escape.prevent="close"
    >
      <div class="pro-modal__backdrop" aria-hidden="true" @click="close" />
      <div
        ref="panelRef"
        class="pro-modal__panel"
        :class="[sizeClass, { 'pro-modal__panel--contain': containScroll }]"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        tabindex="-1"
      >
        <header class="pro-modal__header">
          <h2 :id="titleId" class="pro-modal__title">{{ title }}</h2>
          <button
            type="button"
            class="pro-modal__close"
            :aria-label="resolvedCloseLabel"
            data-testid="pro-modal-close"
            :disabled="preventClose"
            @click="close"
          >
            <ProIcon name="close" :size="22" />
          </button>
        </header>
        <div
          class="pro-modal__body"
          :class="{ 'pro-modal__body--contain': containScroll }"
        >
          <slot />
        </div>
        <footer v-if="$slots.footer" class="pro-modal__footer">
          <slot name="footer" />
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/** Nested modals: keep body scroll locked until the last one closes. */
let modalOpenCount = 0

const props = withDefaults(
  defineProps<{
    open: boolean
    title: string
    closeLabel?: string
    size?: 'md' | 'lg' | 'xl' | 'full'
    /** When true, ignore X / Escape / backdrop close (e.g. save in flight). */
    preventClose?: boolean
    /**
     * Body does not scroll — child fills height and manages its own overflow
     * (e.g. visit-report panel with pinned meta + actions).
     */
    containScroll?: boolean
    /** data-testid on the root overlay (default keeps existing e2e selectors). */
    testId?: string
  }>(),
  { size: 'md', testId: 'pro-modal', preventClose: false, containScroll: false },
)

const emit = defineEmits<{ 'update:open': [boolean] }>()

const { t } = useI18n()
const panelRef = ref<HTMLElement | null>(null)
const titleId = `pro-modal-title-${useId()}`
const stackDepth = ref(0)

const resolvedCloseLabel = computed(() => props.closeLabel || t('common.cancel'))
const sizeClass = computed(() => {
  switch (props.size) {
    case 'full':
      return 'pro-modal__panel--full'
    case 'xl':
      return 'pro-modal__panel--xl'
    case 'lg':
      return 'pro-modal__panel--lg'
    case 'md':
      return 'pro-modal__panel--md'
    default: {
      const _exhaustive: never = props.size
      return _exhaustive
    }
  }
})

function close() {
  if (props.preventClose) return
  emit('update:open', false)
}

watch(
  () => props.open,
  async (isOpen, wasOpen) => {
    if (!import.meta.client) return
    if (isOpen && !wasOpen) {
      modalOpenCount += 1
      stackDepth.value = modalOpenCount
      document.body.style.overflow = 'hidden'
      await nextTick()
      panelRef.value?.focus()
    }
    else if (!isOpen && wasOpen) {
      modalOpenCount = Math.max(0, modalOpenCount - 1)
      stackDepth.value = 0
      if (modalOpenCount === 0) {
        document.body.style.overflow = ''
      }
    }
  },
)

onBeforeUnmount(() => {
  if (!import.meta.client) return
  if (props.open) {
    modalOpenCount = Math.max(0, modalOpenCount - 1)
    if (modalOpenCount === 0) document.body.style.overflow = ''
  }
})
</script>

<style scoped>
.pro-modal {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.25rem;
}

.pro-modal__backdrop {
  position: absolute;
  inset: 0;
  background: color-mix(in srgb, var(--pf-vet-primary) 45%, transparent);
}

.pro-modal__panel {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  width: min(100%, 28rem);
  max-height: min(90vh, 40rem);
  overflow: hidden;
  background: var(--pf-vet-surface);
  border-radius: var(--pf-vet-radius);
  box-shadow: var(--pf-vet-shadow-md);
  outline: none;
}

.pro-modal__panel--lg {
  width: min(100%, 42rem);
}

.pro-modal__panel--xl {
  width: min(96vw, 56rem);
  max-height: min(92vh, 52rem);
}

.pro-modal__panel--full {
  width: min(98vw, 90rem);
  height: 96vh;
  max-height: 96vh;
}

@media (max-width: 720px) {
  .pro-modal:has(.pro-modal__panel--full) {
    padding: 0.5rem;
  }

  .pro-modal__panel--full {
    width: 100%;
    height: 96vh;
    max-height: 96vh;
    border-radius: var(--pf-vet-radius);
  }
}

.pro-modal__header {
  flex-shrink: 0;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1.25rem 1.25rem 0.5rem;
}

.pro-modal__title {
  margin: 0;
  font-size: 1.125rem;
  color: var(--pf-vet-primary);
}

.pro-modal__close {
  appearance: none;
  border: 0;
  background: transparent;
  color: var(--pf-vet-text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: var(--pf-vet-radius);
  display: inline-flex;
  line-height: 1;
}

.pro-modal__close:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.pro-modal__close:hover:not(:disabled) {
  color: var(--pf-vet-primary);
  background: var(--pf-vet-bg);
}

.pro-modal__body {
  flex: 1 1 auto;
  min-height: 0;
  overflow: auto;
  padding: 0.75rem 1.25rem 1.25rem;
}

.pro-modal__body--contain {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding-bottom: 0.75rem;
}

/* Hauteur définie pour que le child (ex. CR) puisse flex + scroller en interne. */
.pro-modal__panel--contain.pro-modal__panel--md {
  height: min(90vh, 40rem);
  max-height: min(90vh, 40rem);
}

.pro-modal__panel--contain.pro-modal__panel--lg {
  height: min(90vh, 40rem);
  max-height: min(90vh, 40rem);
}

.pro-modal__panel--contain.pro-modal__panel--xl {
  height: min(92vh, 52rem);
  max-height: min(92vh, 52rem);
}

.pro-modal__panel--full .pro-modal__body {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding-bottom: 0.75rem;
}

.pro-modal__footer {
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem 1.25rem;
  border-top: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface);
}
</style>
