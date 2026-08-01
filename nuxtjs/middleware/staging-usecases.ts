/**
 * Use cases : staging|local + rôles admin / commercial / commercial_manager.
 * Pose le layout selon le rôle ; une seule URL `/usecases`.
 */
export default defineNuxtRouteMiddleware(async () => {
  const { isStagingLike } = useAppEnv()
  if (!isStagingLike.value) {
    const role = await resolveProRole()
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }

  const role = await resolveProRole()
  if (applySalesOpsLayout(role)) return
  if (isProRole(role)) return navigateTo(homePathForRole(role))
  return navigateTo('/login')
})
