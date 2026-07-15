export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('pf_token')
  if (to.path === '/login') return
  if (!token.value) return navigateTo('/login')
})
