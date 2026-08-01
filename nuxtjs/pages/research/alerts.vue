<template>
  <div data-testid="research-alerts-page">
    <ProPageHeader
      :title="$t('research.alerts.title')"
      :subtitle="$t('research.alerts.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <ProEmptyState
      v-if="showDisabled"
      data-testid="research-disabled"
      :title="$t('research.disabledTitle')"
    />
    <template v-else>
      <p v-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>
      <ProEmptyState v-else-if="loadError" :title="$t('research.loadError')" data-testid="research-alerts-error" />
      <ProTable
        v-else
        :empty="!items.length"
        :empty-title="$t('research.alerts.empty')"
        data-testid="research-alerts-table"
      >
        <thead>
          <tr>
            <th>{{ $t('research.heatmap.postalCode') }}</th>
            <th>{{ $t('research.heatmap.species') }}</th>
            <th>{{ $t('research.heatmap.signal') }}</th>
            <th>{{ $t('research.alerts.week') }}</th>
            <th>{{ $t('research.heatmap.count') }}</th>
            <th>{{ $t('research.alerts.baseline') }}</th>
            <th>{{ $t('research.alerts.zScore') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, i) in items" :key="`${row.postalCode}-${row.signalType}-${i}`">
            <td>{{ row.postalCode || '—' }}</td>
            <td>{{ row.species }}</td>
            <td>{{ row.signalType }}</td>
            <td>{{ row.week }}</td>
            <td>{{ row.eventCount }}</td>
            <td>{{ row.baselineAvg.toFixed(1) }}</td>
            <td>{{ row.zScore.toFixed(2) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </template>
  </div>
</template>

<script setup lang="ts">
import { isPublicFlagOn } from '~/utils/public-feature-flag'

definePageMeta({
  layout: 'research',
  middleware: ['research-only'],
})

const runtimeConfig = useRuntimeConfig()
const apiDisabled = ref(false)
const showDisabled = computed(
  () => !isPublicFlagOn(runtimeConfig.public.researchEnabled) || apiDisabled.value,
)

type AlertRow = {
  postalCode: string
  species: string
  signalType: string
  week: string
  eventCount: number
  baselineAvg: number
  zScore: number
}

const items = ref<AlertRow[]>([])
const loading = ref(false)
const loadError = ref(false)

async function load() {
  if (!isPublicFlagOn(runtimeConfig.public.researchEnabled)) return
  loading.value = true
  loadError.value = false
  apiDisabled.value = false
  try {
    const res: any = await $fetch('/api/research/alerts')
    const data = res?.data ?? res
    items.value = Array.isArray(data?.items) ? data.items : []
  } catch (e: any) {
    const code = e?.data?.error?.code
    if (code === 'research_disabled' || e?.statusCode === 404) {
      apiDisabled.value = true
      items.value = []
    } else {
      loadError.value = true
      items.value = []
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>
