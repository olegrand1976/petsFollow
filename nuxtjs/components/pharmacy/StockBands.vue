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
