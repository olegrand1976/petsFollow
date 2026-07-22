export function useNavBadges() {
  const clientsBadge = useState<number>('nav-clients-badge', () => 0)
  const calendarBadge = useState<number>('nav-calendar-badge', () => 0)

  async function refresh() {
    try {
      const res: any = await $fetch('/api/vet/overview')
      const data = res.data ?? res ?? {}
      clientsBadge.value = Number(data.pendingLinkRequests ?? 0)
      calendarBadge.value = Number(data.pendingVisits ?? 0)
    } catch {
      clientsBadge.value = 0
      calendarBadge.value = 0
    }
  }

  return { clientsBadge, calendarBadge, refresh }
}
