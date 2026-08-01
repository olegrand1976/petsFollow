/** Stable display order for Pro profile switcher buttons. */
const PROFILE_ROLE_ORDER: string[] = [
  'admin',
  'dev',
  'commercial_manager',
  'commercial',
  'vet',
  'vet_assistant',
  'secretary',
  'research',
]

/**
 * Whether a user who owns `ownedRoles` may activate `target`.
 * Mirrors go/pkg/kernel.CanActivateProfile (owned profiles, not active role).
 */
export function canActivateProfile(ownedRoles: string[], target: string): boolean {
  const ownsAdmin = ownedRoles.includes('admin')
  const ownsManager = ownedRoles.includes('commercial_manager')
  const ownsCommercial = ownedRoles.includes('commercial')
  if (ownsAdmin) return true
  if (ownsManager) return target !== 'admin'
  if (ownsCommercial) return target !== 'admin' && target !== 'commercial_manager'
  return true
}

export function profileRoleSortIndex(role: string): number {
  const i = PROFILE_ROLE_ORDER.indexOf(role)
  return i === -1 ? PROFILE_ROLE_ORDER.length : i
}
