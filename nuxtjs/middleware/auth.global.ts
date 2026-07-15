const PUBLIC_PATHS = new Set([
  '/',
  '/login',
  '/register',
  '/register/sent',
  '/confirm-email',
  '/welcome',
])

export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('pf_token')
  const isPublic = PUBLIC_PATHS.has(to.path) || to.path.startsWith('/register')

  if (isPublic) {
    if (token.value && to.path === '/') {
      return navigateTo('/dashboard')
    }
    return
  }

  if (!token.value) return navigateTo('/login')
})
