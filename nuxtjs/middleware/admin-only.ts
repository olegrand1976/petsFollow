export default defineNuxtRouteMiddleware(async () => {
  const role = await resolveProRole()
  if (role !== 'admin') {
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }
})
