const SKIP_PATHS = new Set([
  '/',
  '/produits',
  '/login',
  '/register',
  '/register/sent',
  '/confirm-email',
  '/forgot-password',
  '/reset-password',
  '/change-password',
  '/welcome',
  '/onboarding',
])

export default defineNuxtRouteMiddleware(async (to) => {
  const token = useCookie('pf_token')
  if (!token.value) return
  if (
    SKIP_PATHS.has(to.path)
    || to.path.startsWith('/register')
    || to.path.startsWith('/legal')
    || to.path.startsWith('/admin')
    || to.path.startsWith('/commercial')
    || to.path.startsWith('/commercial-manager')
  ) {
    return
  }

  try {
    const { fetchUser } = useProUser()
    const me = await fetchUser()
    if (!me) {
      clearAuthTokens()
      return navigateTo('/login')
    }
    if (me.mustChangePassword === true) {
      // Force-change takes priority over onboarding.
      return
    }
    if (me.role === 'vet' && me.profileComplete === false && to.path !== '/onboarding') {
      return navigateTo('/onboarding')
    }
  } catch {
    clearAuthTokens()
    return navigateTo('/login')
  }
})
