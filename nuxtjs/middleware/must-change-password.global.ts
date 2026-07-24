const SKIP_PREFIXES = [
  '/change-password',
  '/login',
  '/register',
  '/confirm-email',
  '/forgot-password',
  '/reset-password',
  '/welcome',
  '/legal',
  '/invite',
]

function isUnauthorized(e: unknown): boolean {
  const err = e as { statusCode?: number, status?: number, response?: { status?: number } }
  const status = err?.statusCode ?? err?.status ?? err?.response?.status
  return status === 401 || status === 403
}

export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === '/' || SKIP_PREFIXES.some((p) => to.path === p || to.path.startsWith(`${p}/`))) {
    return
  }
  if (!hasSessionCookie()) return

  try {
    const { fetchUser } = useProUser()
    const me = await fetchUser()
    if (!me) return
    if (me.mustChangePassword === true && to.path !== '/change-password') {
      return navigateTo('/change-password')
    }
    if (me.mustChangePassword !== true && to.path === '/change-password') {
      return navigateTo(homePathForRole(me.role, { profileComplete: me.profileComplete }))
    }
  } catch (e) {
    if (isUnauthorized(e)) {
      await clearAuthTokens()
      return navigateTo('/login')
    }
    // 5xx / network: do not force logout.
  }
})
