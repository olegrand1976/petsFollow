/** Accès réservé au véto de référence du cabinet (ex. /commissions). */
export default defineNuxtRouteMiddleware(async () => {
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
