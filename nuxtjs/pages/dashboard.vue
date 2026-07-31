<template>
  <div data-testid="vet-dashboard-page">
    <ProPageHeader
      :title="welcomeTitle"
      :subtitle="$t('dashboard.subtitle')"
    />
    <div class="pro-grid-kpi">
      <ProKpi
        v-if="canReadClients"
        icon="group"
        :value="clientCount"
        :label="$t('dashboard.activeClients')"
        to="/clients"
      />
      <ProKpi
        v-if="canMessage"
        icon="chat"
        :value="unreadCount"
        :label="$t('dashboard.unreadMessages')"
        to="/messages"
        :variant="hasUnread ? 'alert' : 'default'"
      />
      <ProKpi
        v-if="canReadPets"
        icon="favorite"
        :value="unreadHeartrate"
        :label="$t('dashboard.unreadHeartrate')"
        to="/pets"
        :variant="unreadHeartrateRaw > 0 ? 'alert' : 'default'"
      />
      <ProKpi
        v-if="canManageShares"
        icon="inbox"
        :value="pendingLinks"
        :label="$t('dashboard.pendingLinks')"
        to="/clients?invitations=1"
        :variant="pendingLinksRaw > 0 ? 'alert' : 'default'"
      />
      <ProKpi
        v-if="canManageCalendar"
        icon="event"
        :value="pendingVisits"
        :label="$t('dashboard.pendingVisits')"
        to="/calendar"
        :variant="pendingVisitsRaw > 0 ? 'alert' : 'default'"
      />
    </div>
    <div class="pro-grid-2 pro-mt-lg">
      <ProCard v-if="canManageCalendar" :title="$t('dashboard.todayTitle')" data-testid="dashboard-today">
        <ProEmptyState
          v-if="!todayVisits.length"
          :title="$t('dashboard.todayEmptyTitle')"
          :description="$t('dashboard.todayEmptyDescription')"
        />
        <ul v-else class="pro-dashboard-list">
          <li v-for="v in todayVisits" :key="v.id" class="pro-dashboard-list__item">
            <NuxtLink :to="`/calendar?visit=${v.id}`" class="pro-dashboard-list__link">
              <span class="pro-dashboard-list__main">
                <strong>{{ formatVisitTime(v) }}</strong>
                <span>{{ v.clientName }}<template v-if="v.petName"> · {{ v.petName }}</template></span>
              </span>
              <ProBadge :variant="statusVariant(v.status)">{{ $t(`calendar.status.${v.status}`) }}</ProBadge>
            </NuxtLink>
          </li>
        </ul>
      </ProCard>

      <ProCard :title="$t('dashboard.queueTitle')" data-testid="dashboard-queue">
        <ProEmptyState
          v-if="!queueItems.length"
          :title="$t('dashboard.queueEmptyTitle')"
          :description="$t('dashboard.queueEmptyDescription')"
        />
        <ul v-else class="pro-dashboard-list">
          <li v-for="item in queueItems" :key="item.key" class="pro-dashboard-list__item">
            <NuxtLink :to="item.to" class="pro-dashboard-list__link">
              <ProIcon :name="item.icon" :size="20" class="pro-dashboard-list__icon" />
              <span class="pro-dashboard-list__main">
                <strong>{{ item.label }}</strong>
                <span v-if="item.meta">{{ item.meta }}</span>
              </span>
              <ProBadge v-if="item.badge" variant="warning">{{ item.badge }}</ProBadge>
            </NuxtLink>
          </li>
        </ul>
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CalendarVisit } from '~/composables/useCalendarGrid'

definePageMeta({ middleware: 'vet-only' })

const { t } = useI18n()
const { canPractice } = usePracticePerms()
const canReadClients = computed(() => canPractice('clients.read'))
const canReadPets = computed(() => canPractice('pets.read'))
const canManageShares = computed(() => canPractice('shares.manage'))
const canManageCalendar = computed(() => canPractice('calendar.manage'))
const canMessage = computed(() => canPractice('messaging'))
const welcomeTitle = ref(t('dashboard.title'))
const clientCount = ref('—')
const unreadCount = ref('—')
const unreadHeartrate = ref('—')
const pendingLinks = ref('—')
const pendingVisits = ref('—')
const unreadRaw = ref(0)
const pendingLinksRaw = ref(0)
const pendingVisitsRaw = ref(0)
const unreadHeartrateRaw = ref(0)
const { fetchUser } = useProUser()
const { formatTime } = useFormatters()
const { statusVariant } = useCalendarGrid()

const hasUnread = computed(() => unreadRaw.value > 0)

const todayVisits = ref<CalendarVisit[]>([])
const pendingCalendarVisits = ref<CalendarVisit[]>([])
const unreadThreads = ref<any[]>([])

type QueueItem = {
  key: string
  to: string
  icon: string
  label: string
  meta?: string
  badge?: string
}

function formatVisitTime(v: CalendarVisit) {
  const at = v.scheduledAt || v.proposedScheduledAt
  return at ? formatTime(at) : t('calendar.unscheduled')
}

function threadClientLabel(thread: any) {
  return thread.clientName || t('common.clientFallback', { id: thread.clientUserId?.slice(0, 8) ?? '' })
}

const queueItems = computed<QueueItem[]>(() => {
  const items: QueueItem[] = []
  if (canManageCalendar.value) {
    for (const v of pendingCalendarVisits.value) {
      items.push({
        key: `visit-${v.id}`,
        to: `/calendar?visit=${v.id}`,
        icon: 'event',
        label: v.clientName
          ? `${v.clientName}${v.petName ? ' · ' + v.petName : ''}`
          : t('calendar.unscheduled'),
        meta: t(`calendar.status.${v.status}`),
      })
    }
  }
  if (canMessage.value) {
    for (const th of unreadThreads.value) {
      items.push({
        key: `thread-${th.id}`,
        to: `/messages?thread=${th.id}`,
        icon: 'chat',
        label: threadClientLabel(th),
        meta: th.lastMessagePreview || undefined,
        badge: String(th.unreadCount),
      })
    }
  }
  if (canManageShares.value && pendingLinksRaw.value > 0) {
    items.push({
      key: 'link-requests',
      to: '/clients?invitations=1',
      icon: 'mail',
      label: pendingLinksRaw.value > 1
        ? t('dashboard.queueLinkRequestsPlural', { count: pendingLinksRaw.value })
        : t('dashboard.queueLinkRequests', { count: pendingLinksRaw.value }),
    })
  }
  return items
})

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

async function loadCalendar() {
  if (!canManageCalendar.value) return
  const day0 = new Date()
  day0.setHours(0, 0, 0, 0)
  const day1 = new Date(day0)
  day1.setDate(day1.getDate() + 1)
  try {
    const res: any = await $fetch(
      `/api/vet/calendar?from=${encodeURIComponent(toLocalRFC3339(day0))}&to=${encodeURIComponent(toLocalRFC3339(day1))}`,
    )
    const cal = res.data ?? res
    todayVisits.value = [...(cal.visits ?? [])].sort((a: CalendarVisit, b: CalendarVisit) => {
      const aAt = a.scheduledAt || a.proposedScheduledAt || ''
      const bAt = b.scheduledAt || b.proposedScheduledAt || ''
      return aAt.localeCompare(bAt)
    })
    pendingCalendarVisits.value = (cal.pending ?? []).slice(0, 8)
  } catch { /* ignore */ }
}

async function loadUnreadThreads() {
  if (!canMessage.value) return
  try {
    const res: any = await $fetch('/api/messaging/threads')
    const list = Array.isArray(res?.data ?? res) ? (res.data ?? res) : []
    unreadThreads.value = list
      .filter((th: any) => (th?.unreadCount ?? 0) > 0)
      .sort((a: any, b: any) => {
        const aTime = a.lastMessageAt ? new Date(a.lastMessageAt).getTime() : 0
        const bTime = b.lastMessageAt ? new Date(b.lastMessageAt).getTime() : 0
        return bTime - aTime
      })
      .slice(0, 5)
  } catch { /* ignore */ }
}

onMounted(async () => {
  try {
    const me = await fetchUser()
    const name = me?.fullName
    if (name) welcomeTitle.value = t('dashboard.welcome', { name: name.split(' ')[0] })
  } catch { /* ignore */ }

  // Overview is gated by clients.read on the API — skip if unavailable.
  if (canReadClients.value) {
    try {
      const res: any = await $fetch('/api/vet/overview')
      const data = res.data ?? res
      clientCount.value = String(data.clientCount ?? 0)
      if (canMessage.value) {
        unreadRaw.value = Number(data.unreadMessages ?? 0)
        unreadCount.value = String(unreadRaw.value)
      }
      if (canReadPets.value) {
        unreadHeartrateRaw.value = Number(data.unreadHeartrate ?? 0)
        unreadHeartrate.value = String(unreadHeartrateRaw.value)
      }
      if (canManageShares.value) {
        pendingLinksRaw.value = Number(data.pendingLinkRequests ?? 0)
        pendingLinks.value = String(pendingLinksRaw.value)
      }
      if (canManageCalendar.value) {
        pendingVisitsRaw.value = Number(data.pendingVisits ?? 0)
        pendingVisits.value = String(pendingVisitsRaw.value)
      }
    } catch { /* ignore */ }
  }

  await Promise.all([loadCalendar(), loadUnreadThreads()])
})
</script>

<style scoped>
.pro-dashboard-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
  max-height: 22rem;
  overflow-y: auto;
}

.pro-dashboard-list__link {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem 0.75rem;
  border-radius: var(--pf-vet-radius);
  background: var(--pf-vet-bg);
  color: inherit;
  text-decoration: none;
}

.pro-dashboard-list__link:hover {
  background: color-mix(in srgb, var(--pf-vet-accent) 10%, var(--pf-vet-bg));
}

.pro-dashboard-list__icon {
  flex: none;
  color: var(--pf-vet-text-muted);
}

.pro-dashboard-list__main {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  flex: 1 1 auto;
  min-width: 0;
}

.pro-dashboard-list__main span {
  font-size: 0.8125rem;
  color: var(--pf-vet-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
