<template>
  <svg
    :viewBox="`0 0 ${width} ${height}`"
    class="pro-bpm-chart"
    role="img"
    :aria-label="ariaLabel"
  >
    <polyline
      :points="polyline"
      fill="none"
      stroke="var(--pf-vet-accent)"
      stroke-width="2"
      stroke-linejoin="round"
    />
    <circle
      v-for="(p, i) in plotted"
      :key="i"
      :cx="p.x"
      :cy="p.y"
      r="3"
      :fill="p.alert ? 'var(--pf-vet-alert)' : 'var(--pf-vet-accent)'"
    />
  </svg>
</template>

<script setup lang="ts">
const props = defineProps<{
  values: number[]
  alerts?: boolean[]
  ariaLabel?: string
}>()

const width = 320
const height = 120
const padding = 12

const plotted = computed(() => {
  const vals = props.values
  if (!vals.length) return []
  const max = Math.max(...vals, 80)
  const min = Math.min(...vals, 40)
  const range = Math.max(max - min, 20)
  const step = vals.length > 1 ? (width - padding * 2) / (vals.length - 1) : 0
  return vals.map((v, i) => ({
    x: padding + i * step,
    y: height - padding - ((v - min) / range) * (height - padding * 2),
    alert: props.alerts?.[i] ?? false,
  }))
})

const polyline = computed(() => plotted.value.map(p => `${p.x},${p.y}`).join(' '))
</script>

<style scoped>
.pro-bpm-chart {
  width: 100%;
  max-width: 24rem;
  height: auto;
}
</style>
