<template>
  <div class="pro-bpm-chart-wrap">
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
      <g v-for="(p, i) in plotted" :key="i">
        <circle
          :cx="p.x"
          :cy="p.y"
          r="3.5"
          :fill="p.alert ? 'var(--pf-vet-alert)' : 'var(--pf-vet-accent)'"
        />
        <text
          v-if="p.showLabel"
          :x="p.x"
          :y="p.y - 8"
          text-anchor="middle"
          class="pro-bpm-chart-label"
          :fill="p.alert ? 'var(--pf-vet-alert)' : 'var(--pf-vet-primary)'"
        >
          {{ p.value }}
        </text>
      </g>
    </svg>
    <ul class="pro-bpm-chart-legend" :aria-label="$t('clients.pet.chartLegend')">
      <li>
        <span class="pro-bpm-chart-swatch pro-bpm-chart-swatch--ok" aria-hidden="true" />
        {{ $t('clients.pet.chartLegendOk') }}
      </li>
      <li>
        <span class="pro-bpm-chart-swatch pro-bpm-chart-swatch--alert" aria-hidden="true" />
        {{ $t('clients.pet.chartLegendAlert') }}
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  values: number[]
  alerts?: boolean[]
  ariaLabel?: string
}>()

const width = 360
const height = 148
const paddingX = 16
const paddingTop = 28
const paddingBottom = 14
/** Min horizontal gap (viewBox units) between BPM labels to avoid overlap. */
const minLabelGap = 28

const plotted = computed(() => {
  const vals = props.values
  if (!vals.length) return []
  const max = Math.max(...vals, 80)
  const min = Math.min(...vals, 40)
  const range = Math.max(max - min, 20)
  const step = vals.length > 1 ? (width - paddingX * 2) / (vals.length - 1) : 0
  const plotH = height - paddingTop - paddingBottom
  const points = vals.map((v, i) => ({
    x: paddingX + i * step,
    y: height - paddingBottom - ((v - min) / range) * plotH,
    value: v,
    alert: props.alerts?.[i] ?? false,
    showLabel: false,
  }))

  let lastLabelX = -Infinity
  const lastIdx = points.length - 1
  for (let i = 0; i < points.length; i++) {
    const p = points[i]
    const force = i === 0 || i === lastIdx || p.alert
    if (force || p.x - lastLabelX >= minLabelGap) {
      p.showLabel = true
      lastLabelX = p.x
    }
  }
  // Ensure the last point keeps a label even if a nearby alert stole the slot.
  if (points.length && !points[lastIdx].showLabel) {
    points[lastIdx].showLabel = true
  }
  return points
})

const polyline = computed(() => plotted.value.map(p => `${p.x},${p.y}`).join(' '))
</script>

<style scoped>
.pro-bpm-chart-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
}

.pro-bpm-chart {
  width: 100%;
  max-width: 28rem;
  height: auto;
}

.pro-bpm-chart-label {
  font-size: 9px;
  font-weight: 600;
  font-family: inherit;
}

.pro-bpm-chart-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: 0.8125rem;
  color: var(--pf-vet-text-muted);
}

.pro-bpm-chart-legend li {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.pro-bpm-chart-swatch {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.pro-bpm-chart-swatch--ok {
  background: var(--pf-vet-accent);
}

.pro-bpm-chart-swatch--alert {
  background: var(--pf-vet-alert);
}
</style>
