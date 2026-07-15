<template>
  <div data-testid="pet-detail-page">
    <nav class="pro-breadcrumb" :aria-label="$t('common.breadcrumb')">
      <NuxtLink to="/clients">{{ $t('nav.clients') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <NuxtLink :to="`/clients/${clientId}`">{{ $t('common.profile') }}</NuxtLink>
      <span class="pro-breadcrumb-sep">/</span>
      <span>{{ pet?.name || $t('clients.pet.title') }}</span>
    </nav>
    <ProPageHeader
      :title="pet?.name || $t('clients.pet.title')"
      :subtitle="petSubtitle"
    />

    <div v-if="recentAlert" class="pro-alert-banner" role="status" data-testid="pet-alert-banner">
      <ProBadge variant="danger">{{ $t('clients.pet.alert') }}</ProBadge>
      <span>{{ $t('clients.pet.recentAlertBanner') }}</span>
    </div>

    <ProCard v-if="chartValues.length" :title="$t('clients.pet.chartTitle')">
      <ProBpmChart
        :values="chartValues"
        :alerts="chartAlerts"
        :aria-label="$t('clients.pet.chartTitle')"
      />
    </ProCard>

    <ProCard :title="$t('clients.pet.heartrateTitle')">
      <div class="pro-toggle pro-pet-filter" role="group" :aria-label="$t('clients.pet.heartrateTitle')">
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': sessionFilter === 'all' }"
          data-testid="pet-filter-all"
          @click="sessionFilter = 'all'"
        >
          {{ $t('clients.pet.filterAll') }}
        </button>
        <button
          type="button"
          class="pro-toggle-btn"
          :class="{ 'pro-toggle-btn--active': sessionFilter === 'alerts' }"
          data-testid="pet-filter-alerts"
          @click="sessionFilter = 'alerts'"
        >
          {{ $t('clients.pet.filterAlerts') }}
        </button>
      </div>
      <ProTable
        :empty="!filteredSessions.length"
        :empty-title="$t('clients.pet.heartrateEmptyTitle')"
        :empty-description="$t('clients.pet.heartrateEmptyDescription')"
      >
        <thead>
          <tr>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.columnBpm') }}</th>
            <th>{{ $t('clients.pet.columnDuration') }}</th>
            <th>{{ $t('clients.pet.columnStatus') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="s in filteredSessions"
            :key="s.id"
            :class="{ 'pro-table-row--alert': s.isAlert }"
          >
            <td>{{ formatDate(s.startedAt) }}</td>
            <td><code>{{ s.bpm }}</code></td>
            <td>{{ s.durationSec }}s</td>
            <td>
              <ProBadge :variant="s.isAlert ? 'danger' : 'success'">
                {{ s.isAlert ? $t('clients.pet.alert') : $t('clients.pet.ok') }}
              </ProBadge>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('clients.pet.timelineTitle')">
      <ul v-if="timeline.length" class="pro-timeline">
        <li v-for="item in timeline" :key="item.id" class="pro-timeline__item">
          <div class="pro-timeline__dot" aria-hidden="true" />
          <div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.body }}</p>
            <small class="text-muted">{{ formatDate(item.createdAt) }}</small>
          </div>
        </li>
      </ul>
      <ProEmptyState
        v-else
        :title="$t('clients.pet.timelineEmptyTitle')"
        :description="$t('clients.pet.timelineEmptyDescription')"
      />
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const route = useRoute()
const clientId = route.params.clientId as string
const petId = route.params.petId as string
const pet = ref<any>(null)
const sessions = ref<any[]>([])
const timeline = ref<any[]>([])
const sessionFilter = ref<'all' | 'alerts'>('all')

const { formatDate } = useFormatters()

const petSubtitle = computed(() => {
  if (!pet.value) return ''
  return [pet.value.species, pet.value.breed].filter(Boolean).join(' · ')
})

const filteredSessions = computed(() => {
  if (sessionFilter.value === 'alerts') {
    return sessions.value.filter((s) => s.isAlert)
  }
  return sessions.value
})

const recentAlert = computed(() => sessions.value.length > 0 && !!sessions.value[0]?.isAlert)

const chartSessions = computed(() => [...sessions.value].slice(0, 30).reverse())
const chartValues = computed(() => chartSessions.value.map(s => s.bpm as number).filter(v => v != null))
const chartAlerts = computed(() => chartSessions.value.map(s => !!s.isAlert))

onMounted(async () => {
  const petRes: any = await $fetch(`/api/pets/${petId}`)
  pet.value = petRes.data ?? petRes

  const sessionsRes: any = await $fetch(`/api/pets/${petId}/heartrate`)
  sessions.value = sessionsRes.data ?? sessionsRes ?? []

  const timelineRes: any = await $fetch(`/api/pets/${petId}/timeline`)
  timeline.value = timelineRes.data ?? timelineRes ?? []
})
</script>

<style scoped>
.pro-alert-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  margin-bottom: 1rem;
  border-radius: var(--pf-vet-radius);
  border: 1px solid color-mix(in srgb, var(--pf-vet-alert) 35%, transparent);
  background: color-mix(in srgb, var(--pf-vet-alert) 8%, var(--pf-vet-surface));
}

.pro-pet-filter {
  margin-bottom: 1rem;
}

.pro-table-row--alert {
  background: color-mix(in srgb, var(--pf-vet-alert) 6%, transparent);
}

.pro-timeline {
  list-style: none;
  margin: 0;
  padding: 0;
}

.pro-timeline__item {
  display: grid;
  grid-template-columns: 1rem 1fr;
  gap: 0.75rem 1rem;
  padding-bottom: 1.25rem;
  border-left: 2px solid var(--pf-vet-border);
  margin-left: 0.35rem;
  padding-left: 1.25rem;
  position: relative;
}

.pro-timeline__item:last-child {
  border-left-color: transparent;
  padding-bottom: 0;
}

.pro-timeline__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--pf-vet-accent);
  position: absolute;
  left: -6px;
  top: 0.35rem;
}

.pro-timeline__item p {
  margin: 0.25rem 0;
}
</style>
