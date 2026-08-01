<template>
  <div data-testid="research-heatmap-page">
    <ProPageHeader
      :title="$t('research.heatmap.title')"
      :subtitle="$t('research.heatmap.subtitle')"
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
      <div class="pro-filters pro-mb-md">
        <ProInput v-model="species" :label="$t('research.filters.species')" test-id="research-filter-species" />
        <ProInput v-model="signal" :label="$t('research.filters.signal')" test-id="research-filter-signal" />
        <ProButton test-id="research-heatmap-refresh" @click="load">{{ $t('common.refresh') }}</ProButton>
      </div>
      <p v-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>
      <ProEmptyState v-else-if="loadError" :title="$t('research.loadError')" data-testid="research-heatmap-error" />
      <ProTable
        v-else
        :empty="!items.length"
        :empty-title="$t('research.heatmap.empty')"
        data-testid="research-heatmap-table"
      >
        <thead>
          <tr>
            <th>{{ $t('research.heatmap.postalCode') }}</th>
            <th>{{ $t('research.heatmap.country') }}</th>
            <th>{{ $t('research.heatmap.species') }}</th>
            <th>{{ $t('research.heatmap.signal') }}</th>
            <th>{{ $t('research.heatmap.count') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, i) in items" :key="`${row.postalCode}-${row.signalType}-${i}`">
            <td>{{ row.postalCode || '—' }}</td>
            <td>{{ row.countryCode }}</td>
            <td>{{ row.species }}</td>
            <td>{{ row.signalType }}</td>
            <td>{{ row.eventCount }}</td>
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

type Cell = {
  postalCode: string
  countryCode: string
  species: string
  signalType: string
  eventCount: number
}

const species = ref('')
const signal = ref('')
const items = ref<Cell[]>([])
const loading = ref(false)
const loadError = ref(false)

async function load() {
  if (!isPublicFlagOn(runtimeConfig.public.researchEnabled)) return
  loading.value = true
  loadError.value = false
  apiDisabled.value = false
  try {
    const res: any = await $fetch('/api/research/heatmap', {
      query: {
        species: species.value || undefined,
        signal: signal.value || undefined,
      },
    })
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
