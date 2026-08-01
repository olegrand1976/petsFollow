import { parseFlowProfileId } from '~/data/product-flows/catalog'

/**
 * Flux produit : rôles admin / commercial / commercial_manager.
 * Pose le layout selon le rôle ; une seule URL `/flux`.
 * Canonise `?profile=` (invalide / tableau → id connu).
 */
export default defineNuxtRouteMiddleware(async (to, from) => {
  const role = await resolveProRole()
  const allowed =
    role === 'admin' || role === 'commercial' || role === 'commercial_manager'
  if (!allowed) {
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }

  // Query-only client (même path, fullPath différent) : layout déjà posé.
  // Au premier load, from.fullPath === to.fullPath → on pose le layout.
  const isQueryOnlyNav =
    to.path === '/flux' &&
    from.path === '/flux' &&
    to.fullPath !== from.fullPath
  if (!isQueryOnlyNav) {
    applySalesOpsLayout(role)
  }

  const raw = to.query.profile
  const rawStr = Array.isArray(raw) ? raw[0] : raw
  const parsed = parseFlowProfileId(rawStr)

  // Présent mais invalide, ou tableau → canoniser.
  const needsCanon =
    (raw != null && raw !== '' && !parsed) || Array.isArray(raw)
  if (!needsCanon) return

  return navigateTo(
    {
      path: '/flux',
      query: { ...to.query, profile: parsed ?? 'vet' },
    },
    { replace: true },
  )
})
