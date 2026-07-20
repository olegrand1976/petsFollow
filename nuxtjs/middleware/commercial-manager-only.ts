import { parseJwtRole } from '~/composables/useAuth'

export default defineNuxtRouteMiddleware(() => {
  const token = useCookie('pf_token')
  const role = parseJwtRole(token.value)
  if (role !== 'commercial_manager') {
    if (role === 'admin') return navigateTo('/admin')
    if (role === 'commercial') return navigateTo('/commercial')
    if (role === 'vet') return navigateTo('/dashboard')
    return navigateTo('/login')
  }
})
