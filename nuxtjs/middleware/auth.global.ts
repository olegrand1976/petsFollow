const PUBLIC_PATHS = new Set([
  '/',
  '/produits',
  '/login',
  '/register',
  '/register/sent',
  '/confirm-email',
  '/forgot-password',
  '/reset-password',
  '/welcome',
  '/legal/privacy',
  '/legal/terms',
  '/legal/mentions',
])

export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('pf_token')
  const refresh = useCookie('pf_refresh')
  const hasSession = !!(token.value || refresh.value)
  const isPublic = PUBLIC_PATHS.has(to.path)
    || to.path.startsWith('/register')
    || to.path.startsWith('/legal/')

  if (isPublic) {
    if (hasSession && to.path === '/') {
      const role = parseJwtRole(token.value)
      if (role === 'admin') return navigateTo('/admin')
      if (role === 'commercial') return navigateTo('/commercial')
      return navigateTo('/dashboard')
    }
    return
  }

  if (!hasSession) return navigateTo('/login')
})
