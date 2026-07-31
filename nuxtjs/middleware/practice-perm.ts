/** Accès réservé à une capability team (ex. /recommend → clients.write). */
import {
  canPracticeCapability,
  isPracticeCapability,
  type PracticeCapability,
} from '~/composables/usePracticePerms'
import { homePathForRole, isPracticeStaffRole, isProRole } from '~/composables/useAuth'
import { isDeskLockedFlag } from '~/composables/useDeskSession'
import { resolveProRole } from '~/utils/resolveProRole'

function parseCapabilities(raw: unknown[]): PracticeCapability[] {
  const out: PracticeCapability[] = []
  for (const item of raw) {
    if (typeof item === 'string' && isPracticeCapability(item)) {
      out.push(item)
    }
  }
  return out
}

export default defineNuxtRouteMiddleware(async (to) => {
  if (isDeskLockedFlag()) return

  const single =
    typeof to.meta.practicePerm === 'string' && isPracticeCapability(to.meta.practicePerm)
      ? to.meta.practicePerm
      : null
  const anyCaps = Array.isArray(to.meta.practicePermAny)
    ? parseCapabilities(to.meta.practicePermAny as unknown[])
    : []
  // Prefer practicePermAny when both are set (OR entry).
  const caps: PracticeCapability[] = anyCaps.length ? anyCaps : (single ? [single] : [])
  if (!caps.length) return

  const role = await resolveProRole()
  if (!isPracticeStaffRole(role)) {
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }

  const { user, fetchUser } = useProUser()
  // Cache /me ; force uniquement si la map ACL n’a jamais été hydratée (session pré-feature).
  let me = user.value ?? (await fetchUser(false))
  if (me && me.practicePermissions === undefined) {
    me = await fetchUser(true)
  }
  const perms = me?.practicePermissions ?? null
  const staffRole = me?.role ?? role
  const allowed = caps.some((cap) => canPracticeCapability(cap, staffRole, perms))
  if (!allowed) {
    return navigateTo('/dashboard')
  }
})
