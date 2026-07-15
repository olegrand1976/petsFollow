<template>
  <div>
    <ProPageHeader :title="$t('clients.title')" :subtitle="$t('clients.subtitle')" />
    <ProCard>
      <ProListToolbar v-model:view-mode="viewMode">
        <template #filters>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="client-search">{{ $t('clients.search') }}</label>
            <input
              id="client-search"
              v-model="query"
              type="search"
              class="pro-input"
              :placeholder="$t('clients.searchPlaceholder')"
            />
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="pet-filter">{{ $t('clients.petsFilter') }}</label>
            <select id="pet-filter" v-model="petFilter" class="pro-select">
              <option value="all">{{ $t('clients.petsAll') }}</option>
              <option value="none">{{ $t('clients.petsNone') }}</option>
              <option value="with">{{ $t('clients.petsWith') }}</option>
            </select>
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="sort-by">{{ $t('clients.sort') }}</label>
            <select id="sort-by" v-model="sortBy" class="pro-select">
              <option value="name">{{ $t('clients.sortName') }}</option>
              <option value="pets">{{ $t('clients.sortPets') }}</option>
            </select>
          </div>
        </template>
      </ProListToolbar>

      <ProTable
        v-if="viewMode === 'table'"
        :empty="!filtered.length"
        :empty-title="$t('clients.emptyTitle')"
        :empty-description="$t('clients.emptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('clients.columnClient') }}</th>
            <th>{{ $t('clients.columnEmail') }}</th>
            <th>{{ $t('clients.columnPets') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in filtered" :key="c.userId">
            <td>
              <span class="pro-avatar client-avatar" aria-hidden="true">{{ initials(c.fullName) }}</span>
              {{ c.fullName }}
            </td>
            <td>{{ c.email }}</td>
            <td>{{ c.petCount }}</td>
            <td>
              <NuxtLink :to="`/clients/${c.userId}`">{{ $t('common.profile') }}</NuxtLink>
            </td>
          </tr>
        </tbody>
      </ProTable>

      <ProKanban v-else>
        <ProKanbanColumn
          v-for="col in kanbanColumns"
          :key="col.key"
          :title="col.title"
          :count="col.items.length"
          :empty="!col.items.length"
          :empty-title="$t('common.empty')"
        >
          <NuxtLink
            v-for="c in col.items"
            :key="c.userId"
            :to="`/clients/${c.userId}`"
            class="pro-kanban-card"
          >
            <span class="pro-avatar client-avatar">{{ initials(c.fullName) }}</span>
            <strong>{{ c.fullName }}</strong>
            <p class="pro-kanban-card__meta">{{ c.email }}</p>
            <ProBadge variant="neutral">{{ c.petCount }} {{ petLabel(c.petCount) }}</ProBadge>
          </NuxtLink>
        </ProKanbanColumn>
      </ProKanban>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

type ClientRow = {
  userId: string
  email: string
  fullName: string
  petCount: number
}

const { t } = useI18n()
const { compareStrings } = useFormatters()

const clients = ref<ClientRow[]>([])
const query = ref('')
const petFilter = ref<'all' | 'none' | 'with'>('all')
const sortBy = ref<'name' | 'pets'>('name')
const { viewMode } = useListView('pf-clients-view', 'table')

function petLabel(count: number) {
  return count > 1 ? t('common.pets') : t('common.pet')
}

const filtered = computed(() => {
  let list = [...clients.value]
  const q = query.value.trim().toLowerCase()
  if (q) {
    list = list.filter(
      (c) =>
        c.fullName?.toLowerCase().includes(q) ||
        c.email?.toLowerCase().includes(q),
    )
  }
  if (petFilter.value === 'none') list = list.filter((c) => c.petCount === 0)
  if (petFilter.value === 'with') list = list.filter((c) => c.petCount > 0)
  if (sortBy.value === 'name') {
    list.sort((a, b) => compareStrings(a.fullName, b.fullName))
  } else {
    list.sort((a, b) => b.petCount - a.petCount)
  }
  return list
})

const kanbanColumns = computed(() => [
  {
    key: 'none',
    title: t('clients.kanbanNone'),
    items: filtered.value.filter((c) => c.petCount === 0),
  },
  {
    key: 'one',
    title: t('clients.kanbanOne'),
    items: filtered.value.filter((c) => c.petCount === 1),
  },
  {
    key: 'multi',
    title: t('clients.kanbanMulti'),
    items: filtered.value.filter((c) => c.petCount > 1),
  },
])

function initials(name: string) {
  return (name || '?')
    .split(' ')
    .map((p) => p[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

onMounted(async () => {
  const res: any = await $fetch('/api/clients')
  clients.value = res.data ?? res ?? []
})
</script>

<style scoped>
.client-avatar {
  margin-right: 0.5rem;
  vertical-align: middle;
}
</style>
