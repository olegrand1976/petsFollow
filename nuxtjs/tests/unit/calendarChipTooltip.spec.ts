import { describe, expect, it } from 'vitest'
import { calendarChipTooltip, type CalendarVisit } from '../../composables/useCalendarGrid'

describe('calendarChipTooltip', () => {
  const t = (key: string) => key

  it('includes waiting room and preconsult sent', () => {
    const v: CalendarVisit = {
      id: '1',
      status: 'confirmed',
      waitingRoomAt: '2026-08-02T10:00:00Z',
      preconsultStatus: 'pending',
    }
    const tip = calendarChipTooltip(v, t)
    expect(tip).toContain('calendar.waitingRoomTooltip')
    expect(tip).toContain('calendar.preconsultSentTooltip')
  })

  it('prefers urgent over answered', () => {
    const v: CalendarVisit = {
      id: '1',
      status: 'confirmed',
      preconsultStatus: 'submitted',
      preconsultAlert: 'urgent',
    }
    const tip = calendarChipTooltip(v, t)
    expect(tip).toContain('calendar.preconsultUrgentTooltip')
    expect(tip).not.toContain('calendar.preconsultAnsweredTooltip')
  })
})
