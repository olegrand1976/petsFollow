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
})
