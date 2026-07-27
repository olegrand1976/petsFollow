import { describe, expect, it } from 'vitest'
import { groupTimelineByDay, timelineDayKey } from '../../utils/groupTimelineByDay'

describe('timelineDayKey', () => {
  it('formats local calendar day as YYYY-MM-DD', () => {
    expect(timelineDayKey(new Date(2026, 6, 27, 9, 30))).toBe('2026-07-27')
  })

  it('returns empty string for invalid date', () => {
    expect(timelineDayKey('not-a-date')).toBe('')
  })
})

describe('groupTimelineByDay', () => {
  it('groups same-day events and sorts days/items newest first', () => {
    const groups = groupTimelineByDay([
      { id: 'a', createdAt: '2026-07-26T10:00:00' },
      { id: 'b', createdAt: '2026-07-27T08:00:00' },
      { id: 'c', createdAt: '2026-07-27T18:00:00' },
      { id: 'd', createdAt: '2026-07-25T12:00:00' },
    ])

    expect(groups.map(g => g.dayKey)).toEqual(['2026-07-27', '2026-07-26', '2026-07-25'])
    expect(groups[0].items.map(i => i.id)).toEqual(['c', 'b'])
    expect(groups[0].items).toHaveLength(2)
  })

  it('skips invalid dates', () => {
    expect(groupTimelineByDay([
      { id: 'ok', createdAt: '2026-07-27T12:00:00' },
      { id: 'bad', createdAt: 'nope' },
    ])).toHaveLength(1)
  })

  it('returns empty for empty input', () => {
    expect(groupTimelineByDay([])).toEqual([])
  })
})
