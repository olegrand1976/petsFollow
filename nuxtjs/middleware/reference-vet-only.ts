/** Accès réservé au véto de référence du cabinet (ex. /commissions). */
import { homePathForRole, isPracticeStaffRole, isProRole } from '~/composables/useAuth'
import { isDeskLockedFlag } from '~/composables/useDeskSession'
import { resolveProRole } from '~/utils/resolveProRole'

export default defineNuxtRouteMiddleware(async () => {
  if (isDeskLockedFlag()) return
  const role = await resolveProRole()
  if (!isPracticeStaffRole(role)) {
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }
  const { fetchUser } = useProUser()
  const me = await fetchUser(true)
  if (!(me as { isReferenceVet?: boolean } | null)?.isReferenceVet) {
    return navigateTo('/dashboard')
  }
})
