export type TimelineDayItem = {
  id: string
  type?: string
  title?: string
  body?: string
  createdAt: string | Date
  meta?: Record<string, unknown>
}

export type TimelineDayGroup<T extends TimelineDayItem = TimelineDayItem> = {
  dayKey: string
  items: T[]
}

/** Local calendar day key YYYY-MM-DD (avoids UTC shift from toISOString). */
export function timelineDayKey(value: string | Date): string {
  const d = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/**
 * Group timeline items by local calendar day.
 * Days newest-first; within a day, items newest-first.
 */
export function groupTimelineByDay<T extends TimelineDayItem>(items: T[]): TimelineDayGroup<T>[] {
  const byDay = new Map<string, T[]>()
  for (const item of items) {
    const key = timelineDayKey(item.createdAt)
    if (!key) continue
    const list = byDay.get(key)
    if (list) list.push(item)
    else byDay.set(key, [item])
  }

  const groups: TimelineDayGroup<T>[] = []
  for (const [dayKey, dayItems] of byDay) {
    dayItems.sort((a, b) => +new Date(b.createdAt) - +new Date(a.createdAt))
    groups.push({ dayKey, items: dayItems })
  }
  groups.sort((a, b) => (a.dayKey < b.dayKey ? 1 : a.dayKey > b.dayKey ? -1 : 0))
  return groups
}
