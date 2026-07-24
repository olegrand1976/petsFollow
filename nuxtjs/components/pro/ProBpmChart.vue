<template>
  <div class="pro-bpm-chart-wrap">
    <svg
      :viewBox="`0 0 ${width} ${height}`"
      class="pro-bpm-chart"
      role="img"
      :aria-label="ariaLabel"
    >
      <!-- Horizontal grid + Y ticks -->
      <g class="pro-bpm-chart-grid">
        <line
          v-for="tick in yTicks"
          :key="`gy-${tick.value}`"
          :x1="plotLeft"
          :y1="tick.y"
          :x2="plotRight"
          :y2="tick.y"
          class="pro-bpm-chart-gridline"
        />
      </g>

      <!-- Axes -->
      <line
        :x1="plotLeft"
        :y1="plotTop"
        :x2="plotLeft"
        :y2="plotBottom"
        class="pro-bpm-chart-axis"
      />
      <line
        :x1="plotLeft"
        :y1="plotBottom"
        :x2="plotRight"
        :y2="plotBottom"
        class="pro-bpm-chart-axis"
      />

      <!-- Y tick labels + axis title -->
      <text
        :x="12"
        :y="(plotTop + plotBottom) / 2"
        text-anchor="middle"
        dominant-baseline="middle"
        class="pro-bpm-chart-axis-title"
        :transform="`rotate(-90 12 ${(plotTop + plotBottom) / 2})`"
      >
        {{ $t('clients.pet.chartAxisBpm') }}
      </text>
      <text
        v-for="tick in yTicks"
        :key="`yl-${tick.value}`"
        :x="plotLeft - 6"
        :y="tick.y"
        text-anchor="end"
        dominant-baseline="middle"
        class="pro-bpm-chart-tick"
      >
        {{ tick.value }}
      </text>

      <!-- X tick labels -->
      <text
        v-for="(tick, i) in xTicks"
        :key="`xl-${i}`"
        :x="tick.x"
        :y="height - 6"
        text-anchor="middle"
        class="pro-bpm-chart-tick"
      >
        {{ tick.label }}
      </text>

      <polyline
        v-if="plotted.length > 1"
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
  dates?: string[]
  domainStart?: string
  domainEnd?: string
  ariaLabel?: string
}>()

const { dateLocale } = useFormatters()

const width = 420
const height = 200
const plotLeft = 44
const plotRight = width - 12
const plotTop = 28
const plotBottom = height - 28
/** Min horizontal gap (viewBox units) between BPM labels to avoid overlap. */
const minLabelGap = 28

const plotW = plotRight - plotLeft
const plotH = plotBottom - plotTop

/** ~1 year in ms — use month/year ticks for long domains. */
const YEARISH_MS = 300 * 24 * 60 * 60 * 1000

function formatAxisDate(value: Date, longSpan: boolean) {
  if (longSpan) {
    return new Intl.DateTimeFormat(dateLocale(), {
      month: '2-digit',
      year: '2-digit',
    }).format(value)
  }
  return new Intl.DateTimeFormat(dateLocale(), {
    day: '2-digit',
    month: '2-digit',
  }).format(value)
}

const scale = computed(() => {
  const vals = props.values
  const rawMax = vals.length ? Math.max(...vals) : 80
  const rawMin = vals.length ? Math.min(...vals) : 40
  const yMax = Math.ceil(Math.max(rawMax, 80) / 10) * 10
  const yMin = Math.floor(Math.min(rawMin, 40) / 10) * 10
  const yRange = Math.max(yMax - yMin, 20)

  const times = (props.dates ?? []).map(d => +new Date(d))
  const domainStartMs = props.domainStart
    ? +new Date(props.domainStart)
    : (times.length ? Math.min(...times) : Date.now())
  const domainEndMs = props.domainEnd
    ? +new Date(props.domainEnd)
    : (times.length ? Math.max(...times) : Date.now())
  const tSpan = Math.max(domainEndMs - domainStartMs, 1)

  return { yMin, yMax, yRange, domainStartMs, domainEndMs, tSpan }
})

const yTicks = computed(() => {
  const { yMin, yMax, yRange } = scale.value
  const steps = 4
  const ticks: { value: number; y: number }[] = []
  for (let i = 0; i <= steps; i++) {
    const value = Math.round(yMin + (yRange * i) / steps)
    const y = plotBottom - ((value - yMin) / yRange) * plotH
    ticks.push({ value, y })
  }
  return ticks
})

const xTicks = computed(() => {
  const { domainStartMs, tSpan } = scale.value
  const longSpan = tSpan >= YEARISH_MS
  const count = 4
  const ticks: { x: number; label: string }[] = []
  for (let i = 0; i <= count; i++) {
    const t = domainStartMs + (tSpan * i) / count
    const x = plotLeft + (plotW * i) / count
    ticks.push({ x, label: formatAxisDate(new Date(t), longSpan) })
  }
  // Avoid duplicate labels when span is tiny
  return ticks.filter((tick, i, arr) => i === 0 || tick.label !== arr[i - 1]?.label)
})

const plotted = computed(() => {
  const vals = props.values
  if (!vals.length) return []
  const { yMin, yRange, domainStartMs, tSpan } = scale.value
  const points = vals.map((v, i) => {
    const t = props.dates?.[i] ? +new Date(props.dates[i]!) : domainStartMs
    const ratio = Math.min(1, Math.max(0, (t - domainStartMs) / tSpan))
    return {
      x: plotLeft + ratio * plotW,
      y: plotBottom - ((v - yMin) / yRange) * plotH,
      value: v,
      alert: props.alerts?.[i] ?? false,
      showLabel: false,
    }
  })

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
  max-width: 36rem;
  height: auto;
}

.pro-bpm-chart-gridline {
  stroke: var(--pf-vet-border);
  stroke-width: 1;
  stroke-dasharray: 3 3;
}

.pro-bpm-chart-axis {
  stroke: var(--pf-vet-text-muted);
  stroke-width: 1.25;
}

.pro-bpm-chart-tick {
  font-size: 9px;
  fill: var(--pf-vet-text-muted);
  font-family: inherit;
}

.pro-bpm-chart-axis-title {
  font-size: 9px;
  font-weight: 600;
  fill: var(--pf-vet-text-muted);
  font-family: inherit;
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
