<template>
  <div data-testid="research-timeseries-page">
    <ProPageHeader
      :title="$t('research.timeseries.title')"
      :subtitle="$t('research.timeseries.subtitle')"
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
        <ProInput v-model="species" :label="$t('research.filters.species')" test-id="research-ts-species" />
        <ProInput v-model="signal" :label="$t('research.filters.signal')" test-id="research-ts-signal" />
        <ProInput v-model="country" :label="$t('research.filters.country')" test-id="research-ts-country" />
        <ProInput v-model="postalCode" :label="$t('research.filters.postalCode')" test-id="research-ts-postal" />
        <ProButton test-id="research-timeseries-refresh" @click="load">{{ $t('common.refresh') }}</ProButton>
      </div>
      <p v-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>
      <ProEmptyState v-else-if="loadError" :title="$t('research.loadError')" data-testid="research-timeseries-error" />
      <ProTable
        v-else
        :empty="!items.length"
        :empty-title="$t('research.timeseries.empty')"
        data-testid="research-timeseries-table"
      >
        <thead>
          <tr>
            <th>{{ $t('research.timeseries.week') }}</th>
            <th>{{ $t('research.timeseries.count') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in items" :key="row.week">
            <td>{{ row.week }}</td>
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

type Point = { week: string; eventCount: number }

const species = ref('')
const signal = ref('')
const country = ref('')
const postalCode = ref('')
const items = ref<Point[]>([])
const loading = ref(false)
const loadError = ref(false)

async function load() {
  if (!isPublicFlagOn(runtimeConfig.public.researchEnabled)) return
  loading.value = true
  loadError.value = false
  apiDisabled.value = false
  try {
    const res: any = await $fetch('/api/research/timeseries', {
      query: {
        species: species.value || undefined,
        signal: signal.value || undefined,
        country: country.value || undefined,
        postalCode: postalCode.value || undefined,
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
