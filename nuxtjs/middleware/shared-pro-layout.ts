import { isProRole, parseJwtRole } from '~/composables/useAuth'
import { applySalesOpsLayout } from '~/utils/applySalesOpsLayout'

/**
 * Pose le shell Pro selon le rôle pour les pages partagées multi-layouts
 * (`/produits`, `/nouveautes` : vet → default, admin / commercial / manager → shell dédié).
 */
export default defineNuxtRouteMiddleware(async () => {
  let role: string | null = null
  try {
    const { fetchUser } = useProUser()
    const me = await fetchUser(true)
    role = me?.role ?? null
  } catch {
    return navigateTo('/login')
  }
  // Soft fail (5xx) : fallback JWT encore valide pour le layout uniquement.
  if (!role) {
    role = parseJwtRole(useCookie('pf_token').value)
  }
  if (!isProRole(role)) {
    return navigateTo('/login')
  }
  if (!applySalesOpsLayout(role)) {
    setPageLayout('default')
  }
})
