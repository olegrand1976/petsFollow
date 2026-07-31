import { describe, expect, it } from 'vitest'
import { needsContactPhone } from '../../utils/needsContactPhone'

describe('complete-contact-phone gate', () => {
  it('requires phone for commercial without contactPhone', () => {
    expect(needsContactPhone({ role: 'commercial', contactPhone: '' })).toBe(true)
    expect(needsContactPhone({ role: 'commercial_manager', contactPhone: '  ' })).toBe(true)
  })

  it('skips when phone present or other roles', () => {
    expect(needsContactPhone({ role: 'commercial', contactPhone: '0470 12 34 56' })).toBe(false)
    expect(needsContactPhone({ role: 'vet', contactPhone: '' })).toBe(false)
    expect(needsContactPhone({ role: 'commercial', contactPhone: '', mustChangePassword: true })).toBe(false)
  })
})
