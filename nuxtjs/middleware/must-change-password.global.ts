const SKIP_PREFIXES = ['/change-password', '/login', '/register', '/confirm-email', '/forgot-password', '/reset-password', '/welcome', '/produits']

export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === '/' || SKIP_PREFIXES.some((p) => to.path === p || to.path.startsWith(`${p}/`))) {
    return
  }
  const token = useCookie('pf_token')
  if (!token.value) return

  try {
    const res: any = await $fetch('/api/me')
    const me = res.data ?? res
    if (me.mustChangePassword === true && to.path !== '/change-password') {
      return navigateTo('/change-password')
    }
    if (me.mustChangePassword !== true && to.path === '/change-password') {
      if (me.role === 'admin') return navigateTo('/admin')
      if (me.role === 'commercial') return navigateTo('/commercial')
      return navigateTo('/dashboard')
    }
  } catch {
    /* ignore */
  }
})
