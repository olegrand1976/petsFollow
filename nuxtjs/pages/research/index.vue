<template>
  <div data-testid="research-overview-page">
    <ProPageHeader
      :title="$t('research.overview.title')"
      :subtitle="$t('research.overview.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="research-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <ProEmptyState
      v-if="disabled"
      data-testid="research-disabled"
      :title="$t('research.disabledTitle')"
      :description="$t('research.disabledHint')"
    />

    <template v-else>
      <p v-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>
      <ProEmptyState v-else-if="loadError" :title="$t('research.loadError')" />
      <template v-else-if="overview">
        <div class="pro-grid-kpi" data-testid="research-overview-kpi">
          <ProKpi :value="overview.optInPracticeCount" :label="$t('research.overview.optInPractices')" />
          <ProKpi :value="overview.eventCount" :label="$t('research.overview.events')" />
          <ProKpi :value="overview.weekCount" :label="$t('research.overview.weeks')" />
          <ProKpi :value="overview.latestWeek || '—'" :label="$t('research.overview.latestWeek')" />
        </div>

        <div class="pro-grid-2 pro-mt-lg">
          <ProCard :title="$t('research.overview.bySignal')">
            <div class="pro-bar-chart">
              <div
                v-for="(count, signal) in overview.bySignal"
                :key="signal"
                class="pro-bar-row"
              >
                <span>{{ signalLabel(String(signal)) }}</span>
                <div class="pro-bar-track">
                  <div class="pro-bar-fill" :style="{ width: barWidth(count, signalMax) }" />
                </div>
                <span>{{ count }}</span>
              </div>
              <p v-if="!signalMax" class="pro-hint">{{ $t('research.overview.emptySignals') }}</p>
            </div>
          </ProCard>
          <ProCard :title="$t('research.overview.bySpecies')">
            <div class="pro-bar-chart">
              <div
                v-for="(count, species) in overview.bySpecies"
                :key="species"
                class="pro-bar-row"
              >
                <span>{{ species }}</span>
                <div class="pro-bar-track">
                  <div class="pro-bar-fill" :style="{ width: barWidth(count, speciesMax) }" />
                </div>
                <span>{{ count }}</span>
              </div>
              <p v-if="!speciesMax" class="pro-hint">{{ $t('research.overview.emptySpecies') }}</p>
            </div>
          </ProCard>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { isPublicFlagOn } from '~/utils/public-feature-flag'

definePageMeta({
  layout: 'research',
  middleware: ['research-only'],
})

const { t } = useI18n()
const runtimeConfig = useRuntimeConfig()
const disabled = computed(() => !isPublicFlagOn(runtimeConfig.public.researchEnabled))

type Overview = {
  optInPracticeCount: number
  eventCount: number
  weekCount: number
  latestWeek?: string | null
  bySignal: Record<string, number>
  bySpecies: Record<string, number>
}

const overview = ref<Overview | null>(null)
const loading = ref(false)
const loadError = ref(false)

const signalMax = computed(() => Math.max(0, ...Object.values(overview.value?.bySignal || {})))
const speciesMax = computed(() => Math.max(0, ...Object.values(overview.value?.bySpecies || {})))

function barWidth(count: number, max: number) {
  if (!max) return '0%'
  return `${Math.round((count / max) * 100)}%`
}

function signalLabel(signal: string) {
  const key = `research.signals.${signal}`
  const translated = t(key)
  return translated === key ? signal : translated
}

async function load() {
  if (disabled.value) return
  loading.value = true
  loadError.value = false
  try {
    const res: any = await $fetch('/api/research/overview')
    overview.value = (res?.data ?? res) as Overview
  } catch {
    loadError.value = true
    overview.value = null
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>
