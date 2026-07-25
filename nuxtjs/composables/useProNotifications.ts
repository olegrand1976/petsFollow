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
  const threadsState = useState<any[]>('pro-notif-threads', () => [])
  const loadedState = useState<boolean>('pro-notif-loaded', () => false)

  const unreadThreads = computed(() =>
    threadsState.value.filter((thread) => (thread.unreadCount ?? 0) > 0),
  )

  const items = computed<ProNotificationItem[]>(() =>
    unreadThreads.value.map((thread) => ({
      id: thread.id,
      label: thread.clientName || t('common.clientFallback', { id: thread.clientUserId?.slice(0, 8) ?? '' }),
      preview: thread.lastMessagePreview || undefined,
      href: `/messages?thread=${thread.id}`,
    })),
  )

  const count = computed(() =>
    unreadThreads.value.reduce((sum, t) => sum + (t.unreadCount ?? 0), 0),
  )

  async function refresh() {
    try {
      const res: any = await $fetch('/api/messaging/threads')
      const list = Array.isArray(res?.data) ? res.data : Array.isArray(res) ? res : []
      threadsState.value = list.filter((item: any) => item != null && item.id != null)
    } catch {
      threadsState.value = []
    } finally {
      loadedState.value = true
    }
  }

  sharedRefresh = refresh

  async function markAllRead() {
    await $fetch('/api/messaging/threads/read-all', { method: 'POST' })
    threadsState.value = threadsState.value.map((thread) => ({
      ...thread,
      unreadCount: 0,
    }))
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
