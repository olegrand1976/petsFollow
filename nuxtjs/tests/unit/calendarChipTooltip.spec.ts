import { describe, expect, it } from 'vitest'
import { calendarChipTooltip, visitConsultationCta, type CalendarVisit } from '../../composables/useCalendarGrid'

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

  it('includes site name when present', () => {
    const v: CalendarVisit = {
      id: '1',
      status: 'confirmed',
      siteName: 'Antenne Liège',
      waitingRoomAt: '2026-08-02T10:00:00Z',
    }
    const tip = calendarChipTooltip(v, t)
    expect(tip.startsWith('Antenne Liège')).toBe(true)
    expect(tip).toContain('calendar.waitingRoomTooltip')
  })
})

describe('visitConsultationCta', () => {
  const now = new Date('2026-08-02T14:00:00Z')

  it('offers start on a confirmed upcoming RDV with a known client', () => {
    expect(
      visitConsultationCta({ status: 'confirmed', clientId: 'c1', scheduledAt: '2026-08-02T15:00:00Z' }, now),
    ).toBe('start')
  })

  it('offers start while the slot is still running (vet slightly late)', () => {
    expect(
      visitConsultationCta(
        { status: 'confirmed', clientId: 'c1', scheduledAt: '2026-08-02T13:45:00Z', durationMinutes: 30 },
        now,
      ),
    ).toBe('start')
  })

  it('offers view once the slot has ended or the RDV is done', () => {
    expect(
      visitConsultationCta(
        { status: 'confirmed', clientId: 'c1', scheduledAt: '2026-08-02T13:00:00Z', durationMinutes: 30 },
        now,
      ),
    ).toBe('view')
    expect(visitConsultationCta({ status: 'done', clientId: 'c1' }, now)).toBe('view')
    // Le CR reste consultable même sans clientId.
    expect(visitConsultationCta({ status: 'done' }, now)).toBe('view')
  })

  it('resumes a walk-in session regardless of the slot time', () => {
    expect(
      visitConsultationCta(
        { status: 'confirmed', clientId: 'c1', consultationSession: true, scheduledAt: '2026-08-02T09:00:00Z' },
        now,
      ),
    ).toBe('start')
  })

  it('offers nothing before confirmation or after cancellation', () => {
    expect(visitConsultationCta({ status: 'requested', clientId: 'c1' }, now)).toBeNull()
    expect(visitConsultationCta({ status: 'reschedule_pending', clientId: 'c1' }, now)).toBeNull()
    expect(visitConsultationCta({ status: 'cancelled', clientId: 'c1' }, now)).toBeNull()
  })

  it('needs a client to start a consultation', () => {
    expect(visitConsultationCta({ status: 'confirmed', scheduledAt: '2026-08-02T15:00:00Z' }, now)).toBeNull()
  })
})
