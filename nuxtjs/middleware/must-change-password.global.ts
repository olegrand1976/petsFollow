const SKIP_PREFIXES = [
  '/change-password',
  '/login',
  '/register',
  '/confirm-email',
  '/forgot-password',
  '/reset-password',
  '/welcome',
  '/produits',
  '/legal',
]

export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === '/' || SKIP_PREFIXES.some((p) => to.path === p || to.path.startsWith(`${p}/`))) {
    return
  }
  const token = useCookie('pf_token')
  if (!token.value) return

  try {
    const { fetchUser } = useProUser()
    const me = await fetchUser()
    if (!me) {
      clearAuthTokens()
      return navigateTo('/login')
    }
    if (me.mustChangePassword === true && to.path !== '/change-password') {
      return navigateTo('/change-password')
    }
    if (me.mustChangePassword !== true && to.path === '/change-password') {
      return navigateTo(homePathForRole(me.role, { profileComplete: me.profileComplete }))
    }
  } catch {
    clearAuthTokens()
    return navigateTo('/login')
  }
})
