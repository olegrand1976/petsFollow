import { parseJwtRole } from '~/composables/useAuth'

export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('pf_token')
  const role = parseJwtRole(token.value)
  if (role !== 'vet') {
    if (role === 'admin') return navigateTo('/admin')
    return navigateTo('/login')
  }
})
