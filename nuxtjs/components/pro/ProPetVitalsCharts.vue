<template>
  <div
    class="pro-pet-vitals-charts"
    :class="{ 'pro-pet-vitals-charts--grid': layout === 'grid' }"
    data-testid="pet-vitals-charts"
  >
    <ProCard
      v-if="hasChartData"
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
      v-if="hasWeightChartData"
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
      v-if="hasBpChartData"
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
  </div>
</template>

<script setup lang="ts">
export type ChartRange = '3m' | '6m' | '1y'

withDefaults(
  defineProps<{
    layout?: 'stack' | 'grid'
    chartRange: ChartRange
    weightChartRange: ChartRange
    bpChartRange?: ChartRange
    hasChartData: boolean
    hasWeightChartData: boolean
    hasBpChartData?: boolean
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
  }>(),
  {
    layout: 'stack',
    bpChartRange: '3m',
    hasBpChartData: false,
    bpChartSysValues: () => [],
    bpChartDiaValues: () => [],
    bpChartDates: () => [],
    bpDomainStart: '',
    bpDomainEnd: '',
  },
)

const emit = defineEmits<{
  'update:chartRange': [value: ChartRange]
  'update:weightChartRange': [value: ChartRange]
  'update:bpChartRange': [value: ChartRange]
}>()
</script>

<style scoped>
.pro-pet-vitals-charts {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.pro-pet-vitals-charts--grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}

.pro-pet-filter {
  margin-bottom: 1rem;
}

@media (max-width: 900px) {
  .pro-pet-vitals-charts--grid {
    grid-template-columns: 1fr;
  }
}
</style>
