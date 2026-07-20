import { parseJwtRole } from '~/composables/useAuth'

export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('pf_token')
  const role = parseJwtRole(token.value)
  if (role !== 'vet') {
    if (role === 'admin') return navigateTo('/admin')
    if (role === 'commercial_manager') return navigateTo('/commercial-manager')
    if (role === 'commercial') return navigateTo('/commercial')
    return navigateTo('/login')
  }
})
