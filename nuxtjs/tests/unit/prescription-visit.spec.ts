import { describe, expect, it } from 'vitest'
import {
  formatBrusselsDateTime,
  formatPrescriptionVisitLabel,
  isOrphanLinkedVisit,
  visitIdForSave,
  withEnsuredLinkedVisit,
  type PrescriptionVisitOption,
} from '../../utils/prescription-visit'

describe('prescription-visit helpers', () => {
  const base: PrescriptionVisitOption[] = [
    { id: 'v1', scheduledAt: '2026-08-01T10:00:00Z', status: 'completed', hasFinalReport: true },
  ]

  it('withEnsuredLinkedVisit appends orphan when missing', () => {
    const out = withEnsuredLinkedVisit(base, 'gone')
    expect(out).toHaveLength(2)
    expect(out[1]).toEqual({ id: 'gone', orphan: true })
    expect(withEnsuredLinkedVisit(base, 'v1')).toEqual(base)
    expect(withEnsuredLinkedVisit(base, '')).toEqual(base)
  })

  it('visitIdForSave never sends orphans', () => {
    const withOrphan = withEnsuredLinkedVisit(base, 'gone')
    expect(visitIdForSave(withOrphan, 'gone')).toBe('')
    expect(visitIdForSave(withOrphan, 'v1')).toBe('v1')
    expect(visitIdForSave(withOrphan, '')).toBe('')
    expect(isOrphanLinkedVisit(withOrphan, 'gone')).toBe(true)
    expect(isOrphanLinkedVisit(withOrphan, 'v1')).toBe(false)
  })

  it('formatBrusselsDateTime converts UTC to Europe/Brussels wall time', () => {
    // CEST (UTC+2)
    expect(formatBrusselsDateTime('2026-08-01T10:00:00Z')).toBe('2026-08-01 12:00')
    // CET (UTC+1)
    expect(formatBrusselsDateTime('2026-01-15T10:00:00Z')).toBe('2026-01-15 11:00')
    expect(formatBrusselsDateTime('not-a-date')).toBe('—')
  })

  it('formatPrescriptionVisitLabel handles orphan + CR', () => {
    expect(formatPrescriptionVisitLabel({ id: 'x', orphan: true }, 'Indispo')).toBe('Indispo')
    expect(formatPrescriptionVisitLabel(base[0]!)).toBe('2026-08-01 12:00 · completed · CR')
  })
})
