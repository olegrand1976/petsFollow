export type ProNotificationItem = {
  id: string
  label: string
  preview?: string
  href: string
}

export function useProNotifications() {
  const threadsState = useState<any[]>('pro-notif-threads', () => [])
  const loadedState = useState<boolean>('pro-notif-loaded', () => false)

  const items = computed<ProNotificationItem[]>(() =>
    threadsState.value.map((t) => ({
      id: t.id,
      label: t.clientName || `Client ${t.clientUserId?.slice(0, 8) ?? ''}…`,
      preview: t.lastMessagePreview || undefined,
      href: '/messages',
    })),
  )

  const count = computed(() =>
    threadsState.value.reduce((sum, t) => sum + (t.unreadCount ?? 0), 0),
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

  return { items, count, loaded: loadedState, refresh }
}
