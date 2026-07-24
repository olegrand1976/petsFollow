const PUBLIC_PATHS = new Set([
  '/',
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

const AUTH_ENTRY_PATHS = new Set(['/', '/login', '/register', '/register/sent'])

export default defineNuxtRouteMiddleware(async (to) => {
  const token = useCookie('pf_token')
  const refresh = useCookie('pf_refresh')
  const hasSession = !!(token.value || refresh.value)
  const isPublic = PUBLIC_PATHS.has(to.path)
    || to.path.startsWith('/register')
    || to.path.startsWith('/legal/')
    || to.path.startsWith('/invite/')

  if (isPublic) {
    if (hasSession && AUTH_ENTRY_PATHS.has(to.path)) {
      let role = parseJwtRole(token.value)
      if (!role && refresh.value) {
        try {
          const { fetchUser } = useProUser()
          const me = await fetchUser(true)
          role = me?.role ?? null
        } catch {
          role = null
        }
      }
      if (isProRole(role)) return navigateTo(homePathForRole(role))
      clearAuthTokens()
      if (to.path === '/login' || to.path === '/register') return
      return navigateTo('/login')
    }
    return
  }

  if (!hasSession) return navigateTo('/login')
})
