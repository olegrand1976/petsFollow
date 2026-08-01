<template>
  <div data-testid="research-dataroom-page">
    <ProPageHeader
      :title="$t('research.dataroom.title')"
      :subtitle="$t('research.dataroom.subtitle')"
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
    <ProEmptyState
      v-else-if="groupRequired"
      data-testid="research-dataroom-group-required"
      :title="$t('research.dataroom.forbiddenTitle')"
      :description="$t('research.dataroom.forbiddenHint')"
    >
      <NuxtLink to="/research/groups" class="pro-btn">{{ $t('research.dataroom.goGroups') }}</NuxtLink>
    </ProEmptyState>
    <template v-else>
      <div class="pro-filters pro-mb-md">
        <ProInput v-model="species" :label="$t('research.filters.species')" test-id="research-dr-species" />
        <ProInput v-model="signal" :label="$t('research.filters.signal')" test-id="research-dr-signal" />
        <ProInput v-model="country" :label="$t('research.filters.country')" test-id="research-dr-country" />
        <ProInput v-model="postalCode" :label="$t('research.filters.postalCode')" test-id="research-dr-postal" />
        <ProButton test-id="research-dataroom-refresh" @click="load">{{ $t('common.refresh') }}</ProButton>
      </div>
      <p class="pro-hint pro-mb-sm">{{ $t('research.dataroom.kAnonHint', { k: kAnonymity }) }}</p>
      <p v-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>
      <ProEmptyState v-else-if="loadError" :title="$t('research.loadError')" />
      <ProTable
        v-else
        :empty="!items.length"
        :empty-title="$t('research.dataroom.empty')"
        data-testid="research-dataroom-table"
      >
        <thead>
          <tr>
            <th>{{ $t('research.dataroom.week') }}</th>
            <th>{{ $t('research.heatmap.postalCode') }}</th>
            <th>{{ $t('research.heatmap.country') }}</th>
            <th>{{ $t('research.heatmap.species') }}</th>
            <th>{{ $t('research.dataroom.ageBand') }}</th>
            <th>{{ $t('research.heatmap.signal') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, i) in items" :key="`${row.eventWeek}-${i}`">
            <td>{{ row.eventWeek }}</td>
            <td>{{ row.postalCode || '—' }}</td>
            <td>{{ row.countryCode }}</td>
            <td>{{ row.species }}</td>
            <td>{{ row.ageBand }}</td>
            <td>{{ row.signalType }}</td>
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
const groupRequired = ref(false)
const showDisabled = computed(
  () => !isPublicFlagOn(runtimeConfig.public.researchEnabled) || apiDisabled.value,
)

type EventRow = {
  eventWeek: string
  postalCode: string
  countryCode: string
  species: string
  ageBand: string
  signalType: string
}

const species = ref('')
const signal = ref('')
const country = ref('')
const postalCode = ref('')
const items = ref<EventRow[]>([])
const kAnonymity = ref(5)
const loading = ref(false)
const loadError = ref(false)

async function load() {
  if (!isPublicFlagOn(runtimeConfig.public.researchEnabled)) return
  loading.value = true
  loadError.value = false
  apiDisabled.value = false
  groupRequired.value = false
  try {
    const res: any = await $fetch('/api/research/dataroom/events', {
      query: {
        species: species.value || undefined,
        signal: signal.value || undefined,
        country: country.value || undefined,
        postalCode: postalCode.value || undefined,
      },
    })
    const data = res?.data ?? res
    items.value = Array.isArray(data?.items) ? data.items : []
    if (typeof data?.kAnonymity === 'number') kAnonymity.value = data.kAnonymity
  } catch (e: any) {
    const code = e?.data?.error?.code
    if (code === 'research_disabled' || e?.statusCode === 404) {
      apiDisabled.value = true
    } else if (code === 'research_dataroom_forbidden') {
      groupRequired.value = true
    } else {
      loadError.value = true
    }
    items.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>
