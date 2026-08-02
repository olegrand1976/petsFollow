import { isPracticeStaffRole } from './useAuth'

export type ProNotificationItem = {
  id: string
  label: string
  preview?: string
  href: string
}

const POLL_MS = 8_000

let sharedTimer: ReturnType<typeof setInterval> | null = null
let sharedListeners = 0
let sharedRefresh: (() => Promise<void>) | null = null

function onVisibility() {
  if (document.visibilityState === 'visible') {
    void sharedRefresh?.()
  }
}

export function useProNotifications() {
  const { t } = useI18n()
  const { user } = useProUser()
  const threadsState = useState<any[]>('pro-notif-threads', () => [])
  const deskAlertsState = useState<any[]>('pro-notif-desk-alerts', () => [])
  const loadedState = useState<boolean>('pro-notif-loaded', () => false)
  const canFetchDeskAlerts = computed(() => isPracticeStaffRole(user.value?.role))

  const unreadThreads = computed(() =>
    threadsState.value.filter((thread) => (thread.unreadCount ?? 0) > 0),
  )

  const deskItems = computed<ProNotificationItem[]>(() =>
    deskAlertsState.value.map((alert) => {
      const payload = alert.payload || {}
      const client = payload.clientName || ''
      const pet = payload.petName || ''
      const visitId = payload.visitId || ''
      return {
        id: `desk-${alert.id}`,
        label: t('calendar.waitingRoomNotif'),
        preview: [client, pet].filter(Boolean).join(' · ') || t('calendar.waitingRoomNotifPreview'),
        href: visitId ? `/calendar?visit=${encodeURIComponent(visitId)}` : '/calendar',
      }
    }),
  )

  const items = computed<ProNotificationItem[]>(() => [
    ...deskItems.value,
    ...unreadThreads.value.map((thread) => ({
      id: thread.id,
      label: thread.clientName || t('common.clientFallback', { id: thread.clientUserId?.slice(0, 8) ?? '' }),
      preview: thread.lastMessagePreview || undefined,
      href: `/messages?thread=${thread.id}`,
    })),
  ])

  const count = computed(
    () =>
      deskAlertsState.value.length
      + unreadThreads.value.reduce((sum, th) => sum + (th.unreadCount ?? 0), 0),
  )

  async function refresh() {
    try {
      const deskFetch = canFetchDeskAlerts.value
        ? $fetch('/api/vet/desk-alerts').catch(() => null)
        : Promise.resolve(null)
      const [threadsRes, deskRes]: any[] = await Promise.all([
        $fetch('/api/messaging/threads').catch(() => null),
        deskFetch,
      ])
      const list = Array.isArray(threadsRes?.data)
        ? threadsRes.data
        : Array.isArray(threadsRes)
          ? threadsRes
          : []
      threadsState.value = list.filter((item: any) => item != null && item.id != null)
      if (!canFetchDeskAlerts.value) {
        deskAlertsState.value = []
      } else {
        const alerts = Array.isArray(deskRes?.data)
          ? deskRes.data
          : Array.isArray(deskRes)
            ? deskRes
            : []
        deskAlertsState.value = alerts.filter((a: any) => a != null && a.id != null)
      }
    } catch {
      threadsState.value = []
      deskAlertsState.value = []
    } finally {
      loadedState.value = true
    }
  }

  sharedRefresh = refresh

  async function markAllRead() {
    const jobs: Promise<unknown>[] = [
      $fetch('/api/messaging/threads/read-all', { method: 'POST' }).catch(() => null),
    ]
    if (canFetchDeskAlerts.value) {
      jobs.push($fetch('/api/vet/desk-alerts/read', { method: 'POST', body: {} }).catch(() => null))
    }
    await Promise.all(jobs)
    threadsState.value = threadsState.value.map((thread) => ({
      ...thread,
      unreadCount: 0,
    }))
    deskAlertsState.value = []
  }

  function startPolling() {
    sharedListeners += 1
    if (sharedTimer) return
    void refresh()
    sharedTimer = setInterval(() => {
      if (import.meta.client && document.visibilityState === 'hidden') return
      void sharedRefresh?.()
    }, POLL_MS)
    if (import.meta.client) {
      document.addEventListener('visibilitychange', onVisibility)
    }
  }

  function stopPolling() {
    sharedListeners = Math.max(0, sharedListeners - 1)
    if (sharedListeners > 0) return
    if (sharedTimer) {
      clearInterval(sharedTimer)
      sharedTimer = null
    }
    if (import.meta.client) {
      document.removeEventListener('visibilitychange', onVisibility)
    }
  }

  return { items, count, loaded: loadedState, refresh, markAllRead, startPolling, stopPolling }
}
