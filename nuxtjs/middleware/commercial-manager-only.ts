export default defineNuxtRouteMiddleware(async () => {
  const role = await resolveProRole()
  if (role !== 'commercial_manager') {
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }
})
