<template>
  <div data-testid="research-groups-page">
    <ProPageHeader
      :title="$t('research.groups.title')"
      :subtitle="$t('research.groups.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning">{{ $t('nav.tagDev') }}</ProBadge>
      </template>
    </ProPageHeader>

    <ProEmptyState
      v-if="disabled"
      data-testid="research-disabled"
      :title="$t('research.disabledTitle')"
    />
    <template v-else>
      <ProCard :title="$t('research.groups.createTitle')" class="pro-mb-md">
        <div class="pro-filters">
          <ProInput v-model="newName" :label="$t('research.groups.name')" test-id="research-group-name" />
          <ProInput v-model="newDesc" :label="$t('research.groups.description')" test-id="research-group-desc" />
          <ProButton :loading="creating" test-id="research-group-create" @click="create">
            {{ $t('research.groups.create') }}
          </ProButton>
        </div>
        <p v-if="createError" class="pro-error" role="alert">{{ createError }}</p>
      </ProCard>

      <p v-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>
      <ProEmptyState v-else-if="loadError" :title="$t('research.loadError')" />
      <ProTable
        v-else
        :empty="!items.length"
        :empty-title="$t('research.groups.empty')"
        data-testid="research-groups-table"
      >
        <thead>
          <tr>
            <th>{{ $t('research.groups.name') }}</th>
            <th>{{ $t('research.groups.members') }}</th>
            <th>{{ $t('research.groups.role') }}</th>
            <th>{{ $t('research.groups.dataroom') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="g in items" :key="g.id" data-testid="research-group-row">
            <td>{{ g.name }}</td>
            <td>{{ g.memberCount }}</td>
            <td>{{ g.memberRole }}</td>
            <td>
              <ProBadge :variant="g.dataroomEnabled ? 'success' : 'neutral'">
                {{ g.dataroomEnabled ? $t('research.groups.dataroomOn') : $t('research.groups.dataroomOff') }}
              </ProBadge>
            </td>
            <td>
              <NuxtLink :to="`/research/groups/${g.id}`" class="pro-link">
                {{ $t('research.groups.open') }}
              </NuxtLink>
            </td>
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

const { t } = useI18n()
const runtimeConfig = useRuntimeConfig()
const disabled = computed(() => !isPublicFlagOn(runtimeConfig.public.researchEnabled))

type Group = {
  id: string
  name: string
  description: string
  memberRole: string
  memberCount: number
  dataroomEnabled: boolean
}

const items = ref<Group[]>([])
const loading = ref(false)
const loadError = ref(false)
const newName = ref('')
const newDesc = ref('')
const creating = ref(false)
const createError = ref('')

async function load() {
  if (disabled.value) return
  loading.value = true
  loadError.value = false
  try {
    const res: any = await $fetch('/api/research/groups')
    const data = res?.data ?? res
    items.value = Array.isArray(data?.items) ? data.items : []
  } catch {
    loadError.value = true
    items.value = []
  } finally {
    loading.value = false
  }
}

async function create() {
  creating.value = true
  createError.value = ''
  try {
    await $fetch('/api/research/groups', {
      method: 'POST',
      body: { name: newName.value, description: newDesc.value },
    })
    newName.value = ''
    newDesc.value = ''
    await load()
  } catch {
    createError.value = t('research.groups.createError')
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  void load()
})
</script>
