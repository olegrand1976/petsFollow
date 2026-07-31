import { isPracticeStaffRole } from './useAuth'

/** Capability keys aligned with go/internal/store.DefaultTeamPermissions. */
export const PRACTICE_CAPABILITIES = [
  'clients.read',
  'clients.write',
  'pets.read',
  'pets.write_clinical',
  'heartrate.validate',
  'messaging',
  'calendar.manage',
  'care.manage',
  'shares.read',
  'shares.manage',
  'pharmacy.read',
  'pharmacy.write',
  'practice.settings',
  'team.manage',
  'commissions.view',
] as const

export type PracticeCapability = (typeof PRACTICE_CAPABILITIES)[number]

const PRACTICE_CAPABILITY_SET = new Set<string>(PRACTICE_CAPABILITIES)

export function isPracticeCapability(value: string): value is PracticeCapability {
  return PRACTICE_CAPABILITY_SET.has(value)
}

/**
 * Caps lecture / messaging autorisées en fail-open si /me n’a pas fourni
 * `practicePermissions`. Pas de calendar.manage ni writes (API reste autoritaire).
 */
const FAIL_OPEN_CAPS = new Set<PracticeCapability>([
  'clients.read',
  'pets.read',
  'shares.read',
  'pharmacy.read',
  'messaging',
])

/** Pure helper — testable without Nuxt setup. */
export function canPracticeCapability(
  capability: PracticeCapability,
  role: string | null | undefined,
  permissions: Record<string, boolean> | null | undefined,
): boolean {
  if (permissions == null) {
    if (!isPracticeStaffRole(role)) return false
    return FAIL_OPEN_CAPS.has(capability)
  }
  return permissions[capability] === true
}

export function usePracticePerms() {
  const { user } = useProUser()

  const permissions = computed(() => user.value?.practicePermissions ?? null)

  function canPractice(capability: PracticeCapability): boolean {
    return canPracticeCapability(capability, user.value?.role, permissions.value)
  }

  return { permissions, canPractice }
}
