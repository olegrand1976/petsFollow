export default defineNuxtRouteMiddleware(async () => {
  const role = await resolveProRole()
  if (role !== 'commercial' && role !== 'commercial_manager') {
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }
  // Keep manager shell when browsing shared /commercial/* portfolio pages.
  if (role === 'commercial_manager') {
    setPageLayout('commercial-manager')
  } else {
    setPageLayout('commercial')
  }
})
