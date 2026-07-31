import { describe, expect, it } from 'vitest'
import { pacsPollIntervalMs } from '../../utils/pacs-poll'

describe('pacsPollIntervalMs', () => {
  it('polls faster while starting', () => {
    expect(pacsPollIntervalMs('starting')).toBe(2000)
  })
  it('polls slowly when ready', () => {
    expect(pacsPollIntervalMs('ready')).toBe(30000)
  })
  it('uses medium interval when offline', () => {
    expect(pacsPollIntervalMs('offline')).toBe(10000)
  })
})
