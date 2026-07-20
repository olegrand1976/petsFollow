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
    >
      <template #actions>
        <div class="pro-pet-header-actions">
          <ProBadge v-if="isPrimaryPractice" variant="success">{{ $t('clients.pet.primaryBadge') }}</ProBadge>
          <ProAvatar :src="pet?.photoUrl" :name="pet?.name || ''" size="lg" />
        </div>
      </template>
    </ProPageHeader>

    <ProCard :title="$t('clients.pet.photoTitle')" class="pro-settings-card">
      <ProAvatarUpload
        v-model="petPhotoUrl"
        :name="pet?.name || ''"
        :upload-url="`/api/pets/${petId}/photo`"
        :label="$t('clients.pet.photoChange')"
        :hint="$t('clients.pet.photoHint')"
        @uploaded="onPetPhotoUploaded"
      />
    </ProCard>

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

    <ProCard :title="$t('clients.pet.careTitle')" class="pro-mb-lg">
      <form class="pro-pet-inline-form" @submit.prevent="createCare">
        <input v-model="careDraft.title" class="pro-input" :placeholder="$t('clients.pet.careTitleField')" required />
        <select v-model="careDraft.type" class="pro-input" :aria-label="$t('clients.pet.careType')">
          <option value="vaccination">{{ $t('clients.pet.careTypeVaccination') }}</option>
          <option value="deworming">{{ $t('clients.pet.careTypeDeworming') }}</option>
          <option value="vet_check">{{ $t('clients.pet.careTypeVetCheck') }}</option>
          <option value="dental">{{ $t('clients.pet.careTypeDental') }}</option>
          <option value="farrier">{{ $t('clients.pet.careTypeFarrier') }}</option>
          <option value="fecal_egg">{{ $t('clients.pet.careTypeFecalEgg') }}</option>
          <option value="custom">{{ $t('clients.pet.careTypeCustom') }}</option>
        </select>
        <ProButton type="submit" :disabled="careBusy">{{ $t('clients.pet.careCreate') }}</ProButton>
      </form>
      <ProEmptyState
        v-if="!careReminders.length"
        :title="$t('clients.pet.careEmptyTitle')"
        :description="$t('clients.pet.careEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.pet.careType') }}</th>
            <th>{{ $t('clients.pet.careTitleField') }}</th>
            <th>{{ $t('clients.pet.careDue') }}</th>
            <th>{{ $t('clients.pet.columnStatus') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in careReminders" :key="c.id">
            <td>{{ careTypeLabel(c.type) }}</td>
            <td>{{ c.title }}</td>
            <td>
              {{ formatDate(c.dueAt) }}
              <ProBadge v-if="isCareOverdue(c)" variant="danger">{{ $t('clients.pet.careOverdue') }}</ProBadge>
            </td>
            <td>{{ c.status }}</td>
            <td>
              <div v-if="c.status === 'pending'" class="pro-flex-gap">
                <ProButton :disabled="careBusy" @click="markCareDone(c.id)">{{ $t('clients.pet.careDone') }}</ProButton>
                <ProButton variant="ghost" :disabled="careBusy" @click="postponeCare(c.id, 7)">{{ $t('clients.pet.carePostpone') }}</ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('clients.pet.visitsTitle')" class="pro-mb-lg">
      <form class="pro-pet-inline-form" @submit.prevent="proposeVisit(false)">
        <input v-model="visitDraft.scheduledAt" class="pro-input" type="datetime-local" :aria-label="$t('clients.pet.visitScheduledAt')" required />
        <input v-model="visitDraft.notes" class="pro-input" :placeholder="$t('clients.pet.visitNotes')" />
        <ProButton type="submit" :disabled="visitBusy" variant="secondary">
          {{ $t('clients.pet.visitPropose') }}
        </ProButton>
        <ProButton type="button" :disabled="visitBusy || !visitDraft.scheduledAt" @click="proposeVisit(true)">
          {{ $t('clients.pet.visitConfirmDirect') }}
        </ProButton>
      </form>
      <ProEmptyState
        v-if="!visits.length"
        :title="$t('clients.pet.visitsEmptyTitle')"
        :description="$t('clients.pet.visitsEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('clients.pet.columnDate') }}</th>
            <th>{{ $t('clients.pet.visitNotes') }}</th>
            <th>{{ $t('clients.pet.visitStatus') }}</th>
            <th>{{ $t('clients.pet.visitSource') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="v in visits" :key="v.id">
            <td>{{ formatDate(v.scheduledAt || v.createdAt) }}</td>
            <td>{{ v.notes || '—' }}</td>
            <td>{{ v.status }}</td>
            <td>
              {{ v.source === 'vet' ? $t('clients.pet.visitSourceVet') : $t('clients.pet.visitSourceClient') }}
            </td>
            <td>
              <div class="pro-flex-gap">
                <ProButton
                  v-if="v.status === 'requested' && v.pendingActionBy === 'vet'"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'confirm')"
                >
                  {{ $t('clients.pet.visitConfirm') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'reschedule_pending' && v.pendingActionBy === 'vet'"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'accept_reschedule')"
                >
                  {{ $t('calendar.acceptReschedule') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'reschedule_pending' && v.pendingActionBy === 'vet'"
                  variant="ghost"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'reject_reschedule')"
                >
                  {{ $t('calendar.rejectReschedule') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'confirmed'"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'done')"
                >
                  {{ $t('clients.pet.visitDone') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'requested' || v.status === 'confirmed' || v.status === 'reschedule_pending'"
                  variant="ghost"
                  :disabled="visitBusy"
                  @click="visitAction(v.id, 'cancel')"
                >
                  {{ $t('clients.pet.visitCancel') }}
                </ProButton>
              </div>
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
const petPhotoUrl = ref('')
const sessions = ref<any[]>([])
const timeline = ref<any[]>([])
const careReminders = ref<any[]>([])
const visits = ref<any[]>([])
const sessionFilter = ref<'all' | 'alerts'>('all')
const careBusy = ref(false)
const visitBusy = ref(false)
const careDraft = reactive({ title: '', type: 'vaccination' })
const visitDraft = reactive({ scheduledAt: '', notes: '' })

const { formatDate } = useFormatters()
const { t } = useI18n()
const { user, fetchUser } = useProUser()

function careTypeLabel(type: string) {
  switch (type) {
    case 'vaccination':
      return t('clients.pet.careTypeVaccination')
    case 'deworming':
      return t('clients.pet.careTypeDeworming')
    case 'vet_check':
      return t('clients.pet.careTypeVetCheck')
    case 'dental':
      return t('clients.pet.careTypeDental')
    case 'farrier':
      return t('clients.pet.careTypeFarrier')
    case 'fecal_egg':
      return t('clients.pet.careTypeFecalEgg')
    case 'custom':
      return t('clients.pet.careTypeCustom')
    default:
      return type
  }
}

const petSubtitle = computed(() => {
  if (!pet.value) return ''
  return [pet.value.species, pet.value.breed].filter(Boolean).join(' · ')
})

const isPrimaryPractice = computed(() => {
  const practiceId = user.value?.practiceId
  return Boolean(practiceId && pet.value?.practiceId && practiceId === pet.value.practiceId)
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

function onPetPhotoUploaded(data: any) {
  pet.value = { ...pet.value, ...data }
  petPhotoUrl.value = data?.photoUrl || petPhotoUrl.value
}

function isCareOverdue(c: any) {
  return c.status === 'pending' && c.dueAt && new Date(c.dueAt).getTime() < Date.now()
}

async function loadCareAndVisits() {
  const [careRes, visitsRes]: any[] = await Promise.all([
    $fetch(`/api/pets/${petId}/care-reminders`),
    $fetch(`/api/pets/${petId}/visits`),
  ])
  careReminders.value = careRes.data ?? careRes ?? []
  visits.value = visitsRes.data ?? visitsRes ?? []
}

async function createCare() {
  careBusy.value = true
  try {
    await $fetch(`/api/pets/${petId}/care-reminders`, {
      method: 'POST',
      body: { title: careDraft.title, type: careDraft.type, dueDays: 30 },
    })
    careDraft.title = ''
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function markCareDone(id: string) {
  careBusy.value = true
  try {
    await $fetch(`/api/care-reminders/${id}/done`, { method: 'POST' })
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function postponeCare(id: string, days: number) {
  careBusy.value = true
  try {
    await $fetch(`/api/care-reminders/${id}/postpone`, { method: 'POST', body: { days } })
    await loadCareAndVisits()
  } finally {
    careBusy.value = false
  }
}

async function proposeVisit(confirmDirect: boolean) {
  if (!visitDraft.scheduledAt) return
  visitBusy.value = true
  try {
    await $fetch(`/api/pets/${petId}/visits`, {
      method: 'POST',
      body: {
        notes: visitDraft.notes,
        confirmDirect,
        scheduledAt: new Date(visitDraft.scheduledAt).toISOString(),
      },
    })
    visitDraft.notes = ''
    visitDraft.scheduledAt = ''
    await loadCareAndVisits()
  } finally {
    visitBusy.value = false
  }
}

async function visitAction(id: string, action: string) {
  visitBusy.value = true
  try {
    await $fetch(`/api/visits/${id}`, { method: 'PATCH', body: { action } })
    await loadCareAndVisits()
  } finally {
    visitBusy.value = false
  }
}

onMounted(async () => {
  await fetchUser()
  const petRes: any = await $fetch(`/api/pets/${petId}`)
  pet.value = petRes.data ?? petRes
  petPhotoUrl.value = pet.value?.photoUrl || ''

  const sessionsRes: any = await $fetch(`/api/pets/${petId}/heartrate`)
  sessions.value = sessionsRes.data ?? sessionsRes ?? []

  const timelineRes: any = await $fetch(`/api/pets/${petId}/timeline`)
  timeline.value = timelineRes.data ?? timelineRes ?? []

  await loadCareAndVisits()
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

.pro-pet-header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.pro-pet-inline-form {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.pro-pet-inline-form .pro-input {
  flex: 1 1 10rem;
  min-width: 8rem;
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
