const PUBLIC_PATHS = new Set([
  '/',
  '/login',
  '/register',
  '/register/sent',
  '/confirm-email',
  '/forgot-password',
  '/reset-password',
  '/welcome',
])

export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('pf_token')
  const refresh = useCookie('pf_refresh')
  const hasSession = !!(token.value || refresh.value)
  const isPublic = PUBLIC_PATHS.has(to.path) || to.path.startsWith('/register')

  if (isPublic) {
    if (hasSession && to.path === '/') {
      return navigateTo('/dashboard')
    }
    return
  }

  if (!hasSession) return navigateTo('/login')
})
