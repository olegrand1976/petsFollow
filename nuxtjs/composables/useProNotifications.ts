export type ProNotificationItem = {
  id: string
  label: string
  preview?: string
  href: string
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
      threadsState.value = res.data ?? res ?? []
    } catch {
      threadsState.value = []
    } finally {
      loadedState.value = true
    }
  }

  async function markAllRead() {
    await $fetch('/api/messaging/threads/read-all', { method: 'POST' })
    threadsState.value = threadsState.value.map((thread) => ({
      ...thread,
      unreadCount: 0,
    }))
  }

  return { items, count, loaded: loadedState, refresh, markAllRead }
}
