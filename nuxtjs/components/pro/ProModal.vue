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
        :class="sizeClass"
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
            @click="close"
          >
            <ProIcon name="close" :size="22" />
          </button>
        </header>
        <div class="pro-modal__body">
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
    size?: 'md' | 'lg'
    /** data-testid on the root overlay (default keeps existing e2e selectors). */
    testId?: string
  }>(),
  { size: 'md', testId: 'pro-modal' },
)

const emit = defineEmits<{ 'update:open': [boolean] }>()

const { t } = useI18n()
const panelRef = ref<HTMLElement | null>(null)
const titleId = `pro-modal-title-${useId()}`
const stackDepth = ref(0)

const resolvedCloseLabel = computed(() => props.closeLabel || t('common.cancel'))
const sizeClass = computed(() => {
  switch (props.size) {
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
  width: min(100%, 28rem);
  max-height: min(90vh, 40rem);
  overflow: auto;
  background: var(--pf-vet-surface);
  border-radius: var(--pf-vet-radius);
  box-shadow: var(--pf-vet-shadow-md);
  outline: none;
}

.pro-modal__panel--lg {
  width: min(100%, 42rem);
}

.pro-modal__header {
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

.pro-modal__close:hover {
  color: var(--pf-vet-primary);
  background: var(--pf-vet-bg);
}

.pro-modal__body {
  padding: 0.75rem 1.25rem 1.25rem;
}

.pro-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0 1.25rem 1.25rem;
}
</style>
