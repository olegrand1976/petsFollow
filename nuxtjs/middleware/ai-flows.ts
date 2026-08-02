import { parseAiFlowProfileId } from '~/data/ai-flows/catalog'

/**
 * Flux IA / automatisations : rôles admin / commercial / commercial_manager.
 * Pose le layout selon le rôle ; une seule URL `/flux-ia`.
 * Canonise `?profile=` (invalide → vet).
 */
export default defineNuxtRouteMiddleware(async (to, from) => {
  const role = await resolveProRole()
  const allowed =
    role === 'admin' || role === 'commercial' || role === 'commercial_manager'
  if (!allowed) {
    if (isProRole(role)) return navigateTo(homePathForRole(role))
    return navigateTo('/login')
  }

  const isQueryOnlyNav =
    to.path === '/flux-ia' &&
    from.path === '/flux-ia' &&
    to.fullPath !== from.fullPath
  if (!isQueryOnlyNav) {
    applySalesOpsLayout(role)
  }

  const raw = to.query.profile
  const rawStr = Array.isArray(raw) ? raw[0] : raw
  const parsed = parseAiFlowProfileId(rawStr)

  if (raw == null || raw === '') {
    return navigateTo(
      { path: '/flux-ia', query: { ...to.query, profile: 'vet' } },
      { replace: true },
    )
  }

  const needsCanon = !parsed || Array.isArray(raw)
  if (!needsCanon) return

  return navigateTo(
    {
      path: '/flux-ia',
      query: { ...to.query, profile: parsed ?? 'vet' },
    },
    { replace: true },
  )
})
