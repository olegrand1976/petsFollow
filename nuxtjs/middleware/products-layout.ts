import { isProRole, parseJwtRole } from '~/composables/useAuth'

/** Pick Pro shell layout for /produits (vet / commercial / admin). */
export default defineNuxtRouteMiddleware(async () => {
  const token = useCookie('pf_token')
  let role = parseJwtRole(token.value)
  if (!role) {
    try {
      const { fetchUser } = useProUser()
      const me = await fetchUser(true)
      role = me?.role ?? null
    } catch {
      role = null
    }
  }
  if (!isProRole(role)) {
    return navigateTo('/login')
  }
  switch (role) {
    case 'commercial':
      setPageLayout('commercial')
      break
    case 'commercial_manager':
      setPageLayout('commercial-manager')
      break
    case 'admin':
      setPageLayout('admin')
      break
    default:
      setPageLayout('default')
  }
})
