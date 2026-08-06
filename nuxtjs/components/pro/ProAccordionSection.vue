<template>
  <details
    ref="detailsEl"
    class="pro-accordion"
    :data-testid="dataTestid"
  >
    <summary class="pro-accordion__summary">
      <span class="pro-accordion__title-row">
        <h2 class="pro-accordion__title">{{ title }}</h2>
        <ProIcon name="expand_more" class="pro-accordion__chevron" :size="20" />
      </span>
      <!-- Inside summary so it stays visible when the block is collapsed. -->
      <p v-if="description" class="pro-accordion__description">
        {{ description }}
      </p>
    </summary>
    <div class="pro-accordion__body">
      <slot />
    </div>
  </details>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    title: string
    description?: string
    /** When true, open the section (never force-closes if the user collapsed it). */
    open?: boolean
    dataTestid?: string
  }>(),
  {
    open: false,
    description: undefined,
    dataTestid: undefined,
  },
)

const detailsEl = ref<HTMLDetailsElement | null>(null)

function ensureOpen() {
  if (props.open && detailsEl.value && !detailsEl.value.open) {
    detailsEl.value.open = true
  }
}

onMounted(ensureOpen)
watch(() => props.open, ensureOpen)
</script>

<style scoped>
.pro-accordion {
  margin-bottom: 1rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: var(--pf-vet-radius, 8px);
  background: var(--pf-vet-surface, #fff);
  box-shadow: var(--pf-vet-shadow-sm);
  overflow: hidden;
}

.pro-accordion__summary {
  list-style: none;
  cursor: pointer;
  padding: 1rem 1.15rem;
  user-select: none;
}

.pro-accordion__summary::-webkit-details-marker {
  display: none;
}

.pro-accordion__summary:focus-visible {
  outline: 2px solid var(--pf-vet-accent);
  outline-offset: -2px;
}

.pro-accordion__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.pro-accordion__title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--pf-vet-text);
}

.pro-accordion__chevron {
  flex-shrink: 0;
  color: var(--pf-vet-text-muted);
  transition: transform 0.15s ease;
}

.pro-accordion[open] .pro-accordion__chevron {
  transform: rotate(180deg);
}

.pro-accordion__description {
  margin: 0.4rem 0 0;
  color: var(--pf-vet-text-muted);
  font-size: 0.9rem;
  line-height: 1.4;
  font-weight: 400;
}

.pro-accordion__body {
  padding: 0 1.15rem 1.15rem;
}
</style>
