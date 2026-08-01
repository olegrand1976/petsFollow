import { describe, expect, it } from 'vitest'
import { canActivateProfile, homeSwitchRole, profileRoleSortIndex } from '../../utils/profile-switch'

describe('canActivateProfile', () => {
  it('lets admin home activate any role', () => {
    expect(canActivateProfile('admin', 'secretary')).toBe(true)
    expect(canActivateProfile('admin', 'admin')).toBe(true)
  })

  it('blocks manager home from admin but allows return to manager', () => {
    expect(canActivateProfile('commercial_manager', 'admin')).toBe(false)
    expect(canActivateProfile('commercial_manager', 'commercial')).toBe(true)
    expect(canActivateProfile('commercial_manager', 'commercial_manager')).toBe(true)
  })

  it('blocks commercial home from admin and manager', () => {
    expect(canActivateProfile('commercial', 'admin')).toBe(false)
    expect(canActivateProfile('commercial', 'commercial_manager')).toBe(false)
    expect(canActivateProfile('commercial', 'secretary')).toBe(true)
  })
})

describe('homeSwitchRole', () => {
  it('picks earliest non-client profile', () => {
    expect(
      homeSwitchRole([
        { id: 'c', role: 'client', createdAt: '2020-01-01T00:00:00Z' },
        { id: 's', role: 'secretary', createdAt: '2024-06-01T00:00:00Z' },
        { id: 'm', role: 'commercial', createdAt: '2024-01-01T00:00:00Z' },
      ]),
    ).toBe('commercial')
  })

  it('keeps API order when createdAt missing', () => {
    expect(
      homeSwitchRole([
        { id: '1', role: 'commercial' },
        { id: '2', role: 'secretary' },
      ]),
    ).toBe('commercial')
  })

  it('tie-breaks equal createdAt by id', () => {
    expect(
      homeSwitchRole([
        { id: 'b', role: 'secretary', createdAt: '2024-01-01T00:00:00Z' },
        { id: 'a', role: 'commercial', createdAt: '2024-01-01T00:00:00Z' },
      ]),
    ).toBe('commercial')
  })
})

describe('profileRoleSortIndex', () => {
  it('orders known roles stably', () => {
    expect(profileRoleSortIndex('admin')).toBeLessThan(profileRoleSortIndex('vet'))
    expect(profileRoleSortIndex('vet')).toBeLessThan(profileRoleSortIndex('secretary'))
  })
})
