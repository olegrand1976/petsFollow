<template>
  <div data-testid="calendar-page">
    <ProPageHeader :title="$t('calendar.title')" :subtitle="$t('calendar.subtitle')">
      <template #actions>
        <NuxtLink to="/settings#calendar" class="pro-btn pro-btn--secondary">
          {{ $t('calendar.openSettings') }}
        </NuxtLink>
      </template>
    </ProPageHeader>

    <p v-if="!clientBookingEnabled" class="pro-inline-feedback" role="status">
      {{ $t('calendar.bookingDisabledHint') }}
    </p>
    <p v-if="actionError" class="pro-inline-feedback pro-inline-feedback--error" role="alert">{{ actionError }}</p>

    <ProCard :title="$t('calendar.pendingTitle')" class="pro-mb-lg">
      <ProEmptyState
        v-if="!pending.length"
        :title="$t('calendar.pendingEmptyTitle')"
        :description="$t('calendar.pendingEmptyDescription')"
      />
      <ProTable v-else>
        <thead>
          <tr>
            <th>{{ $t('calendar.columnClient') }}</th>
            <th>{{ $t('calendar.columnPet') }}</th>
            <th>{{ $t('calendar.columnWhen') }}</th>
            <th>{{ $t('calendar.columnStatus') }}</th>
            <th>{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="v in pending"
            :id="`visit-${v.id}`"
            :key="v.id"
            :data-testid="`visit-request-${v.id}`"
            :class="{ 'calendar-row--focus': focusVisitId === v.id }"
          >
            <td>
              <NuxtLink v-if="v.clientId" :to="`/clients/${v.clientId}`">{{ v.clientName }}</NuxtLink>
              <span v-else>{{ v.clientName }}</span>
            </td>
            <td>{{ v.petName }}</td>
            <td>{{ formatWhen(v) }}</td>
            <td>
              <ProBadge :variant="statusVariant(v.status)">{{ statusLabel(v.status) }}</ProBadge>
            </td>
            <td>
              <div class="pro-flex-gap">
                <ProButton
                  v-if="v.status === 'requested'"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'confirm')"
                >
                  {{ $t('calendar.confirm') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'reschedule_pending'"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'accept_reschedule')"
                >
                  {{ $t('calendar.acceptReschedule') }}
                </ProButton>
                <ProButton
                  v-if="v.status === 'reschedule_pending'"
                  variant="ghost"
                  :disabled="busyId === v.id"
                  @click="act(v.id, 'reject_reschedule')"
                >
                  {{ $t('calendar.rejectReschedule') }}
                </ProButton>
                <ProButton variant="ghost" :disabled="busyId === v.id" @click="act(v.id, 'cancel')">
                  {{ $t('calendar.cancel') }}
                </ProButton>
              </div>
            </td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard :title="$t('calendar.weekTitle')">
      <div class="calendar-nav pro-mb-md">
        <ProButton variant="ghost" @click="shiftWeek(-1)">{{ $t('calendar.prevWeek') }}</ProButton>
        <strong>{{ weekLabel }}</strong>
        <ProButton variant="ghost" @click="shiftWeek(1)">{{ $t('calendar.nextWeek') }}</ProButton>
      </div>
      <ProEmptyState
        v-if="!weekVisits.length"
        :title="$t('calendar.weekEmptyTitle')"
        :description="$t('calendar.weekEmptyDescription')"
      />
      <ul v-else class="calendar-list">
        <li
          v-for="v in weekVisits"
          :id="`visit-${v.id}`"
          :key="v.id"
          class="calendar-list__item"
          :class="{ 'calendar-row--focus': focusVisitId === v.id }"
        >
          <div>
            <strong>{{ formatWhen(v) }}</strong>
            <span class="text-muted"> — {{ v.clientName }} / {{ v.petName }}</span>
            <ProBadge :variant="statusVariant(v.status)" class="calendar-list__badge">
              {{ statusLabel(v.status) }}
            </ProBadge>
          </div>
          <div class="pro-flex-gap">
            <ProButton
              v-if="v.status === 'requested'"
              :disabled="busyId === v.id"
              @click="act(v.id, 'confirm')"
            >
              {{ $t('calendar.confirm') }}
            </ProButton>
            <ProButton
              v-if="v.status === 'confirmed'"
              variant="secondary"
              :disabled="busyId === v.id"
              @click="openReschedule(v)"
            >
              {{ $t('calendar.proposeMove') }}
            </ProButton>
          </div>
        </li>
      </ul>
    </ProCard>

    <ProModal v-model:open="rescheduleOpen" :title="$t('calendar.proposeMove')">
      <form class="pro-form" @submit.prevent="submitReschedule">
        <ProInput
          v-model="rescheduleAt"
          type="datetime-local"
          :label="$t('calendar.newSlot')"
          required
        />
        <div class="create-client-actions">
          <ProButton variant="secondary" type="button" @click="rescheduleOpen = false">
            {{ $t('common.cancel') }}
          </ProButton>
          <ProButton type="submit" :disabled="busyId === rescheduleVisitId">
            {{ $t('calendar.sendPropose') }}
          </ProButton>
        </div>
      </form>
    </ProModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const route = useRoute()
const { t } = useI18n()
const { formatDate } = useFormatters()
const { mapError } = useApiError()

const pending = ref<any[]>([])
const weekVisits = ref<any[]>([])
const clientBookingEnabled = ref(false)
const actionError = ref('')
const busyId = ref('')
const focusVisitId = ref('')
const weekStart = ref(startOfWeek(new Date()))

const rescheduleOpen = ref(false)
const rescheduleVisitId = ref('')
const rescheduleAt = ref('')

const weekLabel = computed(() => {
  const end = new Date(weekStart.value)
  end.setDate(end.getDate() + 6)
  return `${formatDate(weekStart.value.toISOString())} – ${formatDate(end.toISOString())}`
})

function startOfWeek(d: Date) {
  const x = new Date(d)
  const day = x.getDay()
  const diff = day === 0 ? -6 : 1 - day
  x.setDate(x.getDate() + diff)
  x.setHours(0, 0, 0, 0)
  return x
}

function shiftWeek(delta: number) {
  const n = new Date(weekStart.value)
  n.setDate(n.getDate() + delta * 7)
  weekStart.value = n
  load()
}

function formatWhen(v: any) {
  if (v.proposedScheduledAt) {
    return `${formatDate(v.proposedScheduledAt)} (${t('calendar.proposed')})`
  }
  if (v.scheduledAt) return formatDate(v.scheduledAt)
  return formatDate(v.createdAt)
}

function statusLabel(status: string) {
  return t(`calendar.status.${status}` as any) || status
}

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'neutral' {
  switch (status) {
    case 'confirmed':
      return 'success'
    case 'requested':
    case 'reschedule_pending':
      return 'warning'
    case 'cancelled':
      return 'danger'
    default:
      return 'neutral'
  }
}

function toLocalRFC3339(d: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  const offMin = -d.getTimezoneOffset()
  const sign = offMin >= 0 ? '+' : '-'
  const abs = Math.abs(offMin)
  const oh = pad(Math.floor(abs / 60))
  const om = pad(abs % 60)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
    + `T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}${sign}${oh}:${om}`
}

async function load() {
  actionError.value = ''
  const from = toLocalRFC3339(weekStart.value)
  const toDate = new Date(weekStart.value)
  toDate.setDate(toDate.getDate() + 7)
  const to = toLocalRFC3339(toDate)
  try {
    const [calRes, schedRes]: any[] = await Promise.all([
      $fetch(`/api/vet/calendar?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),
      $fetch('/api/vet/schedule'),
    ])
    const cal = calRes.data ?? calRes
    pending.value = cal.pending ?? []
    const pendingIds = new Set((pending.value || []).map((v: any) => v.id))
    weekVisits.value = (cal.visits ?? []).filter(
      (v: any) => (v.scheduledAt || v.proposedScheduledAt) && !pendingIds.has(v.id),
    )
    const sched = schedRes.data ?? schedRes
    clientBookingEnabled.value = !!sched.clientBookingEnabled
  } catch (e: any) {
    actionError.value = mapError(e)
  }
}

async function act(id: string, action: string) {
  busyId.value = id
  actionError.value = ''
  try {
    await $fetch(`/api/visits/${id}`, { method: 'PATCH', body: { action } })
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    busyId.value = ''
  }
}

function openReschedule(v: any) {
  rescheduleVisitId.value = v.id
  rescheduleAt.value = ''
  rescheduleOpen.value = true
}

async function submitReschedule() {
  if (!rescheduleAt.value) return
  const iso = new Date(rescheduleAt.value).toISOString()
  busyId.value = rescheduleVisitId.value
  try {
    await $fetch(`/api/visits/${rescheduleVisitId.value}`, {
      method: 'PATCH',
      body: { action: 'propose_reschedule', proposedScheduledAt: iso },
    })
    rescheduleOpen.value = false
    await load()
  } catch (e: any) {
    actionError.value = mapError(e)
  } finally {
    busyId.value = ''
  }
}

function focusVisit(id: string) {
  focusVisitId.value = id
  nextTick(() => {
    document.getElementById(`visit-${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}

onMounted(async () => {
  await load()
  if (typeof route.query.visit === 'string' && route.query.visit) {
    focusVisit(route.query.visit)
  }
})

watch(
  () => route.query.visit,
  (visit) => {
    if (typeof visit === 'string' && visit) focusVisit(visit)
  },
)
</script>

<style scoped>
.calendar-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}
.calendar-list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.calendar-list__item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--pf-vet-border);
}
.calendar-list__badge {
  margin-left: 0.5rem;
}
.calendar-row--focus {
  background: color-mix(in srgb, var(--pf-vet-accent) 12%, transparent);
}
.create-client-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1rem;
}
</style>
