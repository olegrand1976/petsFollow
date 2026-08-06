export type CalendarVisit = {
  id: string
  status: string
  source?: string
  consultationSession?: boolean
  pendingActionBy?: string
  petId?: string
  clientId?: string
  clientName?: string
  petName?: string
  notes?: string
  scheduledAt?: string
  proposedScheduledAt?: string
  createdAt?: string
  durationMinutes?: number
  addressText?: string
  lat?: number
  lng?: number
  preconsultStatus?: string
  preconsultAlert?: string
  waitingRoomAt?: string | null
  visitTypeId?: string
  visitTypeName?: string
  visitTypeColor?: string
  siteId?: string
  siteName?: string
}

/**
 * CTA consultation du détail RDV : écran « Nouvelle consultation » tant que le
 * RDV est à venir (ou walk-in en cours), consultation en lecture une fois le
 * RDV passé — statut done, ou fin de créneau dépassée (l'agenda garde les RDV
 * passés en statut confirmed).
 */
export function visitConsultationCta(
  v: Pick<CalendarVisit, 'status' | 'clientId' | 'consultationSession' | 'scheduledAt' | 'durationMinutes'>,
  now: Date = new Date(),
): 'start' | 'view' | null {
  if (v.status === 'done') return 'view'
  if (v.status !== 'confirmed') return null
  if (v.consultationSession) return v.clientId ? 'start' : null
  if (v.scheduledAt) {
    const start = new Date(v.scheduledAt)
    if (!Number.isNaN(start.getTime())) {
      const slotEnd = new Date(start.getTime() + (v.durationMinutes || 30) * 60_000)
      if (slotEnd < now) return 'view'
    }
  }
  return v.clientId ? 'start' : null
}

/** Native tooltip for calendar chip tags (preconsult / waiting room). */
export function calendarChipTooltip(
  v: CalendarVisit,
  t: (key: string) => string,
): string {
  const parts: string[] = []
  if (v.siteName) parts.push(v.siteName)
  if (v.waitingRoomAt) parts.push(t('calendar.waitingRoomTooltip'))
  if (v.preconsultAlert === 'urgent') parts.push(t('calendar.preconsultUrgentTooltip'))
  else if (v.preconsultStatus === 'submitted') parts.push(t('calendar.preconsultAnsweredTooltip'))
  else if (v.preconsultStatus === 'pending') parts.push(t('calendar.preconsultSentTooltip'))
  return parts.join(' — ')
}

export type CalendarVacation = {
  id: string
  startsOn: string
  endsOn: string
  label?: string
}

export function useCalendarGrid() {
  function startOfDay(d: Date) {
    const x = new Date(d)
    x.setHours(0, 0, 0, 0)
    return x
  }

  function startOfWeek(d: Date) {
    const x = startOfDay(d)
    const day = x.getDay()
    const diff = day === 0 ? -6 : 1 - day
    x.setDate(x.getDate() + diff)
    return x
  }

  function startOfMonth(d: Date) {
    return startOfDay(new Date(d.getFullYear(), d.getMonth(), 1))
  }

  function dayKey(d: Date) {
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
  }

  function visitDisplayAt(v: CalendarVisit): Date | null {
    const raw = v.proposedScheduledAt || v.scheduledAt || v.createdAt
    if (!raw) return null
    const d = new Date(raw)
    return Number.isNaN(d.getTime()) ? null : d
  }

  function isUnscheduled(v: CalendarVisit) {
    return !v.proposedScheduledAt && !v.scheduledAt
  }

  function visitsByDay(visits: CalendarVisit[]) {
    const map = new Map<string, CalendarVisit[]>()
    for (const v of visits) {
      const at = visitDisplayAt(v)
      if (!at) continue
      const key = dayKey(at)
      const list = map.get(key) ?? []
      list.push(v)
      map.set(key, list)
    }
    for (const list of map.values()) {
      list.sort((a, b) => {
        const ta = visitDisplayAt(a)?.getTime() ?? 0
        const tb = visitDisplayAt(b)?.getTime() ?? 0
        return ta - tb
      })
    }
    return map
  }

  function vacationOnDay(vacations: CalendarVacation[], day: Date) {
    const key = dayKey(day)
    return vacations.find((v) => v.startsOn <= key && key <= v.endsOn) ?? null
  }

  function weekDays(weekStart: Date) {
    return Array.from({ length: 7 }, (_, i) => {
      const d = new Date(weekStart)
      d.setDate(d.getDate() + i)
      return startOfDay(d)
    })
  }

  /** Grille mois : du lundi de la 1re semaine au dimanche de la dernière (≤ 42 jours). */
  function monthGridRange(monthStart: Date) {
    const first = startOfMonth(monthStart)
    const gridStart = startOfWeek(first)
    const last = startOfDay(new Date(first.getFullYear(), first.getMonth() + 1, 0))
    const lastWeekStart = startOfWeek(last)
    const gridEnd = new Date(lastWeekStart)
    gridEnd.setDate(gridEnd.getDate() + 7)
    const days: Date[] = []
    for (let d = new Date(gridStart); d < gridEnd; d.setDate(d.getDate() + 1)) {
      days.push(startOfDay(d))
    }
    return { gridStart, gridEnd, days }
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

  return {
    startOfDay,
    startOfWeek,
    startOfMonth,
    dayKey,
    visitDisplayAt,
    isUnscheduled,
    visitsByDay,
    vacationOnDay,
    weekDays,
    monthGridRange,
    statusVariant,
  }
}
