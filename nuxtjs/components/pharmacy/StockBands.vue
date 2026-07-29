<template>
  <div class="stock-bands" data-testid="stock-bands">
    <button
      v-for="b in cards"
      :key="b.key"
      type="button"
      class="stock-band"
      :class="{ 'stock-band--active': active === b.key }"
      :data-testid="`stock-band-${b.key}`"
      @click="$emit('select', b.key)"
    >
      <span class="stock-band__label">{{ b.label }}</span>
      <strong class="stock-band__n">{{ summary[b.field] ?? 0 }}</strong>
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  cards: { key: string, field: string, label: string }[]
  active: string
  summary: Record<string, number>
}>()
defineEmits<{ select: [key: string] }>()
</script>

<style scoped>
.stock-bands {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(7.5rem, 1fr));
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.stock-band {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
  padding: 0.65rem 0.75rem;
  text-align: left;
  cursor: pointer;
}
.stock-band--active { outline: 2px solid var(--pf-vet-primary); }
.stock-band__label { display: block; font-size: 0.75rem; color: var(--pf-vet-muted); }
.stock-band__n { font-size: 1.25rem; }
</style>
