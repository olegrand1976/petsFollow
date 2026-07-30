import { describe, expect, it } from 'vitest'
import {
  DEFAULT_DESK_IDLE_MINUTES,
  normalizeDeskIdleMinutes,
  resolveDeskIdleMinutes,
} from '../../utils/deskIdleMinutes'

describe('normalizeDeskIdleMinutes', () => {
  it('keeps allowlisted values', () => {
    for (const n of [1, 2, 5, 10, 15, 30]) {
      expect(normalizeDeskIdleMinutes(n)).toBe(n)
    }
  })

  it('falls back to default for invalid values', () => {
    expect(normalizeDeskIdleMinutes(0)).toBe(DEFAULT_DESK_IDLE_MINUTES)
    expect(normalizeDeskIdleMinutes(7)).toBe(DEFAULT_DESK_IDLE_MINUTES)
    expect(normalizeDeskIdleMinutes(60)).toBe(DEFAULT_DESK_IDLE_MINUTES)
    expect(normalizeDeskIdleMinutes('x')).toBe(DEFAULT_DESK_IDLE_MINUTES)
  })
})

describe('resolveDeskIdleMinutes', () => {
  it('prefers server-synced roster over localStorage (shared-desk hardening)', () => {
    expect(resolveDeskIdleMinutes({
      practiceId: 'prac-a',
      roster: { practiceId: 'prac-a', deskIdleMinutes: 2 },
      scopedIdleRaw: '30',
      legacyIdleRaw: '30',
    })).toBe(2)
  })

  it('uses roster value for matching practice', () => {
    expect(resolveDeskIdleMinutes({
      practiceId: 'prac-a',
      roster: { practiceId: 'prac-a', deskIdleMinutes: 5 },
    })).toBe(5)
  })

  it('falls back to scoped localStorage when roster has no deskIdleMinutes', () => {
    expect(resolveDeskIdleMinutes({
      practiceId: 'prac-a',
      roster: { practiceId: 'prac-a' },
      scopedIdleRaw: '10',
    })).toBe(10)
  })

  it('falls back to legacy key when scoped missing', () => {
    expect(resolveDeskIdleMinutes({
      practiceId: 'prac-a',
      roster: { practiceId: 'prac-a' },
      legacyIdleRaw: '15',
    })).toBe(15)
  })

  it('ignores roster from another practice and uses scoped cache', () => {
    expect(resolveDeskIdleMinutes({
      practiceId: 'prac-b',
      roster: { practiceId: 'prac-a', deskIdleMinutes: 1 },
      scopedIdleRaw: '10',
    })).toBe(10)
  })

  it('defaults when nothing is available', () => {
    expect(resolveDeskIdleMinutes({})).toBe(DEFAULT_DESK_IDLE_MINUTES)
  })
})
