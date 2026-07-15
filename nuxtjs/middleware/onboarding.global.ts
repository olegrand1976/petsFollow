const SKIP_PATHS = new Set([
  '/',
  '/login',
  '/register',
  '/register/sent',
  '/confirm-email',
  '/welcome',
  '/onboarding',
])

export default defineNuxtRouteMiddleware(async (to) => {
  const token = useCookie('pf_token')
  if (!token.value) return
  if (SKIP_PATHS.has(to.path) || to.path.startsWith('/register') || to.path.startsWith('/admin')) {
    return
  }

  try {
    const res: any = await $fetch('/api/me')
    const me = res.data ?? res
    if (me.role === 'vet' && me.profileComplete === false && to.path !== '/onboarding') {
      return navigateTo('/onboarding')
    }
  } catch {
    /* ignore */
  }
})
