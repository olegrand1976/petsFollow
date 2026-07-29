export default defineNuxtRouteMiddleware(async () => {
  const role = await resolveProRole()
  if (!isOpsRole(role)) {
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }
})
