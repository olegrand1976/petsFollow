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

function isUnauthorized(e: unknown): boolean {
  const err = e as { statusCode?: number, status?: number, response?: { status?: number } }
  const status = err?.statusCode ?? err?.status ?? err?.response?.status
  return status === 401 || status === 403
}

export default defineNuxtRouteMiddleware(async (to) => {
  const token = useCookie('pf_token')
  const refresh = useCookie('pf_refresh')
  if (!token.value && !refresh.value) return
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
    if (!me) return
    if (me.mustChangePassword === true) {
      // Force-change takes priority over onboarding.
      return
    }
    if (me.role === 'vet' && me.profileComplete === false && to.path !== '/onboarding') {
      return navigateTo('/onboarding')
    }
  } catch (e) {
    if (isUnauthorized(e)) {
      clearAuthTokens()
      return navigateTo('/login')
    }
  }
})
