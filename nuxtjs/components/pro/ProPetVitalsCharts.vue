<template>
  <div
    class="pro-pet-vitals-charts"
    :class="{ 'pro-pet-vitals-charts--grid': layout === 'grid' }"
    data-testid="pet-vitals-charts"
  >
    <ProCard
      v-if="showEmpty || hasChartData"
      :title="$t('clients.pet.chartTitle')"
      data-testid="pet-bpm-chart-card"
    >
      <div class="pro-toggle pro-pet-filter" role="group" :aria-label="$t('clients.pet.chartRangeLabel')">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': chartRange === '3m' }"
          :aria-pressed="chartRange === '3m'"
          data-testid="pet-chart-range-3m"
          @click="emit('update:chartRange', '3m')"
        >
          {{ $t('clients.pet.chartRange3m') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': chartRange === '6m' }"
          :aria-pressed="chartRange === '6m'"
          data-testid="pet-chart-range-6m"
          @click="emit('update:chartRange', '6m')"
        >
          {{ $t('clients.pet.chartRange6m') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': chartRange === '1y' }"
          :aria-pressed="chartRange === '1y'"
          data-testid="pet-chart-range-1y"
          @click="emit('update:chartRange', '1y')"
        >
          {{ $t('clients.pet.chartRange1y') }}
        </button>
      </div>
      <ProBpmChart
        v-if="chartValues.length"
        :values="chartValues"
        :alerts="chartAlerts"
        :dates="chartDates"
        :domain-start="domainStart"
        :domain-end="domainEnd"
        :aria-label="$t('clients.pet.chartTitle')"
      />
      <p v-else class="text-muted" data-testid="pet-chart-empty-period">
        {{ $t('clients.pet.chartEmptyPeriod') }}
      </p>
    </ProCard>

    <ProCard
      v-if="showEmpty || hasWeightChartData"
      :title="$t('clients.pet.weightChartTitle')"
      data-testid="pet-weight-chart-card"
    >
      <div class="pro-toggle pro-pet-filter" role="group" :aria-label="$t('clients.pet.chartRangeLabel')">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': weightChartRange === '3m' }"
          :aria-pressed="weightChartRange === '3m'"
          data-testid="pet-weight-range-3m"
          @click="emit('update:weightChartRange', '3m')"
        >
          {{ $t('clients.pet.chartRange3m') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': weightChartRange === '6m' }"
          :aria-pressed="weightChartRange === '6m'"
          data-testid="pet-weight-range-6m"
          @click="emit('update:weightChartRange', '6m')"
        >
          {{ $t('clients.pet.chartRange6m') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': weightChartRange === '1y' }"
          :aria-pressed="weightChartRange === '1y'"
          data-testid="pet-weight-range-1y"
          @click="emit('update:weightChartRange', '1y')"
        >
          {{ $t('clients.pet.chartRange1y') }}
        </button>
      </div>
      <ProBpmChart
        v-if="weightChartValues.length"
        :values="weightChartValues"
        :dates="weightChartDates"
        :domain-start="weightDomainStart"
        :domain-end="weightDomainEnd"
        :axis-title="$t('clients.pet.weightAxisKg')"
        auto-y-domain
        hide-legend
        :aria-label="$t('clients.pet.weightChartTitle')"
      />
      <p v-else class="text-muted">
        {{ $t('clients.pet.chartEmptyPeriod') }}
      </p>
    </ProCard>

    <ProCard
      v-if="showEmpty || hasBpChartData"
      :title="$t('clients.pet.bpChartTitle')"
      data-testid="pet-bp-chart-card"
    >
      <div class="pro-toggle pro-pet-filter" role="group" :aria-label="$t('clients.pet.chartRangeLabel')">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': bpChartRange === '3m' }"
          :aria-pressed="bpChartRange === '3m'"
          data-testid="pet-bp-range-3m"
          @click="emit('update:bpChartRange', '3m')"
        >
          {{ $t('clients.pet.chartRange3m') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': bpChartRange === '6m' }"
          :aria-pressed="bpChartRange === '6m'"
          data-testid="pet-bp-range-6m"
          @click="emit('update:bpChartRange', '6m')"
        >
          {{ $t('clients.pet.chartRange6m') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': bpChartRange === '1y' }"
          :aria-pressed="bpChartRange === '1y'"
          data-testid="pet-bp-range-1y"
          @click="emit('update:bpChartRange', '1y')"
        >
          {{ $t('clients.pet.chartRange1y') }}
        </button>
      </div>
      <ProBpmChart
        v-if="bpChartSysValues.length"
        :values="bpChartSysValues"
        :secondary-values="bpChartDiaValues"
        :secondary-primary-label="$t('clients.pet.bpSystolic')"
        :secondary-secondary-label="$t('clients.pet.bpDiastolic')"
        :dates="bpChartDates"
        :domain-start="bpDomainStart"
        :domain-end="bpDomainEnd"
        :axis-title="$t('clients.pet.bpAxisMmHg')"
        auto-y-domain
        :aria-label="$t('clients.pet.bpChartTitle')"
      />
      <p v-else class="text-muted">
        {{ $t('clients.pet.chartEmptyPeriod') }}
      </p>
    </ProCard>

    <ProCard
      v-if="showEmpty || hasLabChartData"
      :title="$t('clients.pet.labsChartTitle')"
      data-testid="pet-lab-chart-card"
    >
      <div class="pro-pet-lab-chart-toolbar">
        <select
          :value="labTrendAnalyte"
          class="pro-input"
          data-testid="pet-labs-trend-select"
          :aria-label="$t('clients.pet.labsTrend')"
          @change="onLabAnalyteChange(($event.target as HTMLSelectElement).value)"
        >
          <option value="">{{ $t('clients.pet.labsTrendSelect') }}</option>
          <option v-for="opt in labAnalyteOptions" :key="opt.code" :value="opt.code">
            {{ opt.label }}
          </option>
        </select>
        <div class="pro-toggle pro-pet-filter" role="group" :aria-label="$t('clients.pet.chartRangeLabel')">
          <button
            type="button"
            class="pro-toggle-btn"
            :class="{ 'pro-toggle-btn--active': labChartRange === '3m' }"
            :aria-pressed="labChartRange === '3m'"
            data-testid="pet-lab-range-3m"
            @click="emit('update:labChartRange', '3m')"
          >
            {{ $t('clients.pet.chartRange3m') }}
          </button>
          <button
            type="button"
            class="pro-toggle-btn"
            :class="{ 'pro-toggle-btn--active': labChartRange === '6m' }"
            :aria-pressed="labChartRange === '6m'"
            data-testid="pet-lab-range-6m"
            @click="emit('update:labChartRange', '6m')"
          >
            {{ $t('clients.pet.chartRange6m') }}
          </button>
          <button
            type="button"
            class="pro-toggle-btn"
            :class="{ 'pro-toggle-btn--active': labChartRange === '1y' }"
            :aria-pressed="labChartRange === '1y'"
            data-testid="pet-lab-range-1y"
            @click="emit('update:labChartRange', '1y')"
          >
            {{ $t('clients.pet.chartRange1y') }}
          </button>
        </div>
      </div>
      <ProBpmChart
        v-if="labTrendAnalyte && labChartValues.length"
        :values="labChartValues"
        :dates="labChartDates"
        :domain-start="labDomainStart"
        :domain-end="labDomainEnd"
        auto-y-domain
        hide-legend
        :axis-title="labAxisTitle"
        :aria-label="$t('clients.pet.labsChartTitle')"
      />
      <p v-else class="text-muted">
        {{ labTrendAnalyte ? $t('clients.pet.chartEmptyPeriod') : $t('clients.pet.labsTrendSelect') }}
      </p>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
export type ChartRange = '3m' | '6m' | '1y'

withDefaults(
  defineProps<{
    layout?: 'stack' | 'grid'
    /** When true, always render one card per reading type (empty state). */
    showEmpty?: boolean
    chartRange: ChartRange
    weightChartRange: ChartRange
    bpChartRange?: ChartRange
    labChartRange?: ChartRange
    hasChartData: boolean
    hasWeightChartData: boolean
    hasBpChartData?: boolean
    hasLabChartData?: boolean
    chartValues: number[]
    chartAlerts: boolean[]
    chartDates: string[]
    domainStart: string
    domainEnd: string
    weightChartValues: number[]
    weightChartDates: string[]
    weightDomainStart: string
    weightDomainEnd: string
    bpChartSysValues?: number[]
    bpChartDiaValues?: number[]
    bpChartDates?: string[]
    bpDomainStart?: string
    bpDomainEnd?: string
    labTrendAnalyte?: string
    labAnalyteOptions?: Array<{ code: string; label: string }>
    labChartValues?: number[]
    labChartDates?: string[]
    labDomainStart?: string
    labDomainEnd?: string
    labAxisTitle?: string
  }>(),
  {
    layout: 'stack',
    showEmpty: false,
    bpChartRange: '3m',
    labChartRange: '1y',
    hasBpChartData: false,
    hasLabChartData: false,
    bpChartSysValues: () => [],
    bpChartDiaValues: () => [],
    bpChartDates: () => [],
    bpDomainStart: '',
    bpDomainEnd: '',
    labTrendAnalyte: '',
    labAnalyteOptions: () => [],
    labChartValues: () => [],
    labChartDates: () => [],
    labDomainStart: '',
    labDomainEnd: '',
    labAxisTitle: '',
  },
)

const emit = defineEmits<{
  'update:chartRange': [value: ChartRange]
  'update:weightChartRange': [value: ChartRange]
  'update:bpChartRange': [value: ChartRange]
  'update:labChartRange': [value: ChartRange]
  'update:labTrendAnalyte': [value: string]
}>()

function onLabAnalyteChange(value: string) {
  emit('update:labTrendAnalyte', value)
}
</script>

<style scoped>
.pro-pet-vitals-charts {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.pro-pet-vitals-charts--grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.pro-pet-filter {
  margin-bottom: 1rem;
}

.pro-pet-lab-chart-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 0.75rem 1rem;
  margin-bottom: 0.25rem;
}

.pro-pet-lab-chart-toolbar .pro-input {
  flex: 1 1 12rem;
  min-width: 10rem;
  max-width: 18rem;
}

.pro-pet-lab-chart-toolbar .pro-pet-filter {
  margin-bottom: 0;
}

@media (max-width: 900px) {
  .pro-pet-vitals-charts--grid {
    grid-template-columns: 1fr;
  }
}
</style>
