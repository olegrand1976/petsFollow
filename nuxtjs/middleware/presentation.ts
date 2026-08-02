import { parsePresentationStepId } from '~/data/presentation/catalog'

/**
 * Présentation cabinet : rôles admin / commercial / commercial_manager.
 * Pose le layout selon le rôle ; une seule URL `/presentation`.
 * Canonise `?step=` (invalide / tableau → welcome).
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
    to.path === '/presentation' &&
    from.path === '/presentation' &&
    to.fullPath !== from.fullPath
  if (!isQueryOnlyNav) {
    applySalesOpsLayout(role)
  }

  const raw = to.query.step
  const rawStr = Array.isArray(raw) ? raw[0] : raw
  const parsed = parsePresentationStepId(rawStr)

  // Absente → welcome ; invalide / tableau → welcome.
  if (raw == null || raw === '') {
    return navigateTo(
      { path: '/presentation', query: { ...to.query, step: 'welcome' } },
      { replace: true },
    )
  }

  const needsCanon = !parsed || Array.isArray(raw)
  if (!needsCanon) return

  return navigateTo(
    {
      path: '/presentation',
      query: { ...to.query, step: parsed ?? 'welcome' },
    },
    { replace: true },
  )
})
