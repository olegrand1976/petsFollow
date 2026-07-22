<template>
  <div data-testid="pets-page">
    <ProPageHeader :title="$t('pets.title')" :subtitle="$t('pets.subtitle')" />
    <p v-if="loadError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ loadError }}</p>

    <ProCard>
      <ProListToolbar :show-view-toggle="false">
        <template #filters>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="pet-search">{{ $t('pets.search') }}</label>
            <input
              id="pet-search"
              v-model="query"
              type="search"
              class="pro-input"
              :placeholder="$t('pets.searchPlaceholder')"
              data-testid="pets-search"
            />
          </div>
          <div class="pro-field pro-field-inline">
            <label class="pro-label" for="species-filter">{{ $t('pets.speciesFilter') }}</label>
            <select id="species-filter" v-model="speciesFilter" class="pro-select" data-testid="pets-species-filter">
              <option value="all">{{ $t('pets.speciesAll') }}</option>
              <option value="dog">{{ $t('common.species.dog') }}</option>
              <option value="cat">{{ $t('common.species.cat') }}</option>
              <option value="horse">{{ $t('common.species.horse') }}</option>
            </select>
          </div>
        </template>
      </ProListToolbar>

      <ProTable
        :empty="!filtered.length"
        :empty-title="$t('pets.emptyTitle')"
        :empty-description="$t('pets.emptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('pets.columnName') }}</th>
            <th>{{ $t('pets.columnSpecies') }}</th>
            <th>{{ $t('pets.columnBreed') }}</th>
            <th>{{ $t('pets.columnBirthDate') }}</th>
            <th>{{ $t('pets.columnLastVisit') }}</th>
            <th>{{ $t('pets.columnLastHeartRate') }}</th>
            <th>{{ $t('pets.columnOwner') }}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in filtered" :key="p.id" :data-testid="`pet-row-${p.id}`">
            <td>
              <div class="pets-name-cell">
                <ProAvatar :src="p.photoUrl" :name="p.name" />
                <span>{{ p.name }}</span>
                <ProBadge
                  v-if="(p.unreadHeartrateCount ?? 0) > 0"
                  variant="danger"
                  data-testid="pet-unread-badge"
                >
                  {{ unreadLabel(p.unreadHeartrateCount ?? 0) }}
                </ProBadge>
              </div>
            </td>
            <td>{{ speciesLabel(p.species) }}</td>
            <td>{{ p.breed || $t('common.unknownBreed') }}</td>
            <td>{{ formatBirthDate(p.birthDate) }}</td>
            <td>{{ p.lastVisitAt ? formatDate(p.lastVisitAt) : $t('common.dash') }}</td>
            <td>
              <template v-if="p.lastHeartRateAt">
                <code v-if="p.lastHeartRateBpm != null">{{ p.lastHeartRateBpm }}</code>
                <span class="text-muted"> · {{ formatDate(p.lastHeartRateAt) }}</span>
              </template>
              <template v-else>{{ $t('common.dash') }}</template>
            </td>
            <td>{{ p.ownerName }}</td>
            <td>
              <NuxtLink :to="`/clients/${p.ownerUserId}/pets/${p.id}`" data-testid="pet-fiche-link">
                {{ $t('common.profile') }}
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

type VetPet = {
  id: string
  ownerUserId: string
  ownerName: string
  name: string
  species: string
  breed?: string
  birthDate?: string
  photoUrl?: string
  lastVisitAt?: string
  lastHeartRateAt?: string
  lastHeartRateBpm?: number
  unreadHeartrateCount?: number
}

const { t, te } = useI18n()
const { formatDate } = useFormatters()
const { mapError } = useApiError()
const { petsBadge, refresh: refreshNavBadges } = useNavBadges()

const pets = ref<VetPet[]>([])
const loadError = ref('')
const query = ref('')
const speciesFilter = ref('all')
let pollTimer: ReturnType<typeof setInterval> | null = null

function speciesLabel(species: string) {
  const key = `common.species.${species}`
  return te(key) ? t(key) : species
}

function unreadLabel(count: number) {
  return count > 1 ? t('pets.unreadCount', { n: count }) : t('pets.unreadBadge')
}

function syncPetsBadgeFromList() {
  petsBadge.value = pets.value.reduce((sum, p) => sum + (p.unreadHeartrateCount ?? 0), 0)
}

function formatBirthDate(value?: string) {
  if (!value) return t('common.dash')
  // birthDate may be date-only (YYYY-MM-DD) — avoid timezone shift
  const day = String(value).slice(0, 10)
  if (/^\d{4}-\d{2}-\d{2}$/.test(day)) {
    const [y, m, d] = day.split('-')
    return `${d}/${m}/${y}`
  }
  return formatDate(value)
}

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  return pets.value.filter((p) => {
    if (speciesFilter.value !== 'all' && p.species !== speciesFilter.value) return false
    if (!q) return true
    const hay = [p.name, p.breed, p.ownerName, p.species].filter(Boolean).join(' ').toLowerCase()
    return hay.includes(q)
  })
})

async function loadPets(silent = false) {
  try {
    const res: any = await $fetch('/api/vet/pets')
    pets.value = res.data ?? res ?? []
    syncPetsBadgeFromList()
  } catch (e: any) {
    if (!silent) loadError.value = mapError(e)
  }
}

onMounted(async () => {
  await loadPets()
  await refreshNavBadges()
  syncPetsBadgeFromList()
  pollTimer = setInterval(() => loadPets(true), 8000)
})

onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.pets-name-cell {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}
</style>
