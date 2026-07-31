import { describe, expect, it } from 'vitest'
import { formatAudioClock } from '../../utils/audioDuration'

describe('formatAudioClock', () => {
  it('formats under one hour as MM:SS', () => {
    expect(formatAudioClock(0)).toBe('00:00')
    expect(formatAudioClock(154)).toBe('02:34')
    expect(formatAudioClock(59)).toBe('00:59')
  })

  it('formats one hour and above as H:MM:SS', () => {
    expect(formatAudioClock(3600)).toBe('1:00:00')
    expect(formatAudioClock(3661)).toBe('1:01:01')
  })

  it('floors and clamps negatives', () => {
    expect(formatAudioClock(61.9)).toBe('01:01')
    expect(formatAudioClock(-3)).toBe('00:00')
  })
})
