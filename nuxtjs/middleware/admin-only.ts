import { parseJwtRole } from '~/composables/useAuth'

export default defineNuxtRouteMiddleware(() => {
  const token = useCookie('pf_token')
  const role = parseJwtRole(token.value)
  if (role !== 'admin') {
    if (role === 'vet') return navigateTo('/clients')
    return navigateTo('/login')
  }
})
