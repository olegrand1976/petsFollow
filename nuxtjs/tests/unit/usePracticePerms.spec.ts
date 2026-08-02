import { describe, expect, it } from 'vitest'
import {
  canPracticeCapability,
  isPracticeCapability,
} from '../../composables/usePracticePerms'

describe('canPracticeCapability', () => {
  it('fail-open lecture/messaging pour staff sans map, refuse calendar et writes', () => {
    expect(canPracticeCapability('clients.read', 'vet', null)).toBe(true)
    expect(canPracticeCapability('pets.read', 'secretary', null)).toBe(true)
    expect(canPracticeCapability('shares.read', 'secretary', null)).toBe(true)
    expect(canPracticeCapability('pharmacy.read', 'secretary', null)).toBe(true)
    expect(canPracticeCapability('messaging', 'secretary', null)).toBe(true)
    expect(canPracticeCapability('calendar.manage', 'vet', null)).toBe(false)
    expect(canPracticeCapability('consultations.history.read', 'vet', null)).toBe(false)
    expect(canPracticeCapability('shares.manage', 'vet', null)).toBe(false)
    expect(canPracticeCapability('pharmacy.write', 'vet', null)).toBe(false)
    expect(canPracticeCapability('clients.write', 'vet', null)).toBe(false)
    expect(canPracticeCapability('pets.write_clinical', 'vet', null)).toBe(false)
    expect(canPracticeCapability('care.manage', 'vet', null)).toBe(false)
    expect(canPracticeCapability('practice.settings', 'vet', null)).toBe(false)
    expect(canPracticeCapability('commissions.view', 'vet', null)).toBe(false)
    expect(canPracticeCapability('team.manage', 'vet', null)).toBe(false)
  })

  it('refuse tout si rôle non-staff et map absente', () => {
    expect(canPracticeCapability('clients.read', 'commercial', null)).toBe(false)
    expect(canPracticeCapability('clients.read', undefined, null)).toBe(false)
  })

  it('honore la map effective quand présente', () => {
    const perms = {
      'clients.read': true,
      'commissions.view': false,
      'pets.write_clinical': false,
    }
    expect(canPracticeCapability('clients.read', 'secretary', perms)).toBe(true)
    expect(canPracticeCapability('commissions.view', 'secretary', perms)).toBe(false)
    expect(canPracticeCapability('pets.write_clinical', 'secretary', perms)).toBe(false)
    expect(canPracticeCapability('messaging', 'secretary', perms)).toBe(false)
  })

  it('isPracticeCapability valide le vocabulaire connu', () => {
    expect(isPracticeCapability('clients.read')).toBe(true)
    expect(isPracticeCapability('not.a.cap')).toBe(false)
    expect(isPracticeCapability('')).toBe(false)
  })
})
