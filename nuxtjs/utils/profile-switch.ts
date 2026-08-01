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
 * Home role for the switch matrix: earliest non-client profile (stable across switches).
 * Mirrors store.homeSwitchRole ordering (createdAt ASC).
 */
export function homeSwitchRole(profiles: { role: string; createdAt?: string }[]): string {
  const nonClient = profiles
    .filter((p) => p.role !== 'client')
    .slice()
    .sort((a, b) => String(a.createdAt || '').localeCompare(String(b.createdAt || '')))
  return nonClient[0]?.role || ''
}

/**
 * Whether an account with `homeRole` may activate `target`.
 * Mirrors go/pkg/kernel.CanActivateProfile.
 */
export function canActivateProfile(homeRole: string, target: string): boolean {
  switch (homeRole) {
    case 'admin':
      return true
    case 'commercial_manager':
      return target !== 'admin'
    case 'commercial':
      return target !== 'admin' && target !== 'commercial_manager'
    default:
      return true
  }
}

export function profileRoleSortIndex(role: string): number {
  const i = PROFILE_ROLE_ORDER.indexOf(role)
  return i === -1 ? PROFILE_ROLE_ORDER.length : i
}
