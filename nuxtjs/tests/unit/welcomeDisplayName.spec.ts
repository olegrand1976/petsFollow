import { describe, expect, it } from 'vitest'
import { welcomeDisplayName } from '~/utils/welcome-display-name'

describe('welcomeDisplayName', () => {
  it('skips Dr / Mme prefixes', () => {
    expect(welcomeDisplayName('Dr Martin Demo')).toBe('Martin')
    expect(welcomeDisplayName('Dr. Martin Demo')).toBe('Martin')
    expect(welcomeDisplayName('Mme Sophie Dupont')).toBe('Sophie')
  })

  it('returns first token when no title', () => {
    expect(welcomeDisplayName('Martin Demo')).toBe('Martin')
  })

  it('handles empty / title-only', () => {
    expect(welcomeDisplayName('')).toBe('')
    expect(welcomeDisplayName(null)).toBe('')
    expect(welcomeDisplayName('Dr')).toBe('Dr')
  })
})
