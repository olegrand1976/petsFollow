import { describe, expect, it } from 'vitest'
import { canActivateProfile, profileRoleSortIndex } from '~/utils/profile-switch'

describe('canActivateProfile', () => {
  const adminOwned = ['admin', 'vet', 'secretary', 'commercial', 'commercial_manager', 'research', 'dev']
  const managerOwned = ['commercial_manager', 'commercial', 'vet', 'secretary', 'research', 'dev']
  const commercialOwned = ['commercial', 'vet', 'secretary', 'research', 'dev']

  it('lets admin activate any owned role', () => {
    expect(canActivateProfile(adminOwned, 'secretary')).toBe(true)
    expect(canActivateProfile(adminOwned, 'admin')).toBe(true)
  })

  it('blocks manager from admin but allows return to manager', () => {
    expect(canActivateProfile(managerOwned, 'admin')).toBe(false)
    expect(canActivateProfile(managerOwned, 'commercial')).toBe(true)
    expect(canActivateProfile(managerOwned, 'commercial_manager')).toBe(true)
  })

  it('blocks commercial from admin and manager', () => {
    expect(canActivateProfile(commercialOwned, 'admin')).toBe(false)
    expect(canActivateProfile(commercialOwned, 'commercial_manager')).toBe(false)
    expect(canActivateProfile(commercialOwned, 'secretary')).toBe(true)
  })
})

describe('profileRoleSortIndex', () => {
  it('orders known roles stably', () => {
    expect(profileRoleSortIndex('admin')).toBeLessThan(profileRoleSortIndex('vet'))
    expect(profileRoleSortIndex('vet')).toBeLessThan(profileRoleSortIndex('secretary'))
  })
})
