import {
  AUTH_LOGIN_REASON_PRO_ONLY,
  clearAuthTokens,
  clearAuthTokensUnlessPostLoginGrace,
  hasSessionCookie,
  homePathForRole,
  isProRole,
} from '~/composables/useAuth'
import { isDeskLockedFlag } from '~/composables/useDeskSession'

const PUBLIC_PATHS = new Set([
  '/',
  '/login',
  '/register',
  '/register/sent',
  '/confirm-email',
  '/forgot-password',
  '/reset-password',
  '/welcome',
  '/legal/privacy',
  '/legal/terms',
  '/legal/mentions',
])

const AUTH_ENTRY_PATHS = new Set(['/', '/login', '/register', '/register/sent'])

export default defineNuxtRouteMiddleware(async (to) => {
  const hasSession = hasSessionCookie()
  const isPublic = PUBLIC_PATHS.has(to.path)
    || to.path.startsWith('/register')
    || to.path.startsWith('/legal/')
    || to.path.startsWith('/invite/')
    || to.path.startsWith('/preconsult/')
    || to.path.startsWith('/dossier/')

  if (isPublic) {
    if (hasSession && AUTH_ENTRY_PATHS.has(to.path)) {
      // Toujours valider via /me — un JWT expiré/révoqué ne doit pas rediriger
      // vers le home (boucle SSR login↔dashboard si getCookie reste stale).
      try {
        const { fetchUser } = useProUser()
        const me = await fetchUser(true)
        // Soft fail (5xx/network) : ne pas purger une session encore valide.
        if (!me) return
        if (isProRole(me.role)) {
          const home = homePathForRole(me.role, { profileComplete: me.profileComplete })
          // Locale hors du try auth : un échec setLocale ne doit pas logout.
          try {
            const { applyPreferredLocale } = useLocaleSync()
            await applyPreferredLocale(me.preferredLocale)
          } catch {
            /* best-effort */
          }
          return navigateTo(home)
        }
        // Session OK mais rôle non-Pro (ex. client) — message explicite sur /login.
        await clearAuthTokens()
        return navigateTo({ path: '/login', query: { reason: AUTH_LOGIN_REASON_PRO_ONLY } })
      } catch {
        // 401/403 : ne pas logout pendant la grâce post-login (course cookies, client only).
        const cleared = await clearAuthTokensUnlessPostLoginGrace()
        if (!cleared) return
      }
      // Landing et écrans auth : rester après purge (pas de re-redirect).
      if (to.path === '/' || to.path === '/login' || to.path === '/register') return
      return navigateTo('/login')
    }
    return
  }

  if (!hasSessionCookie()) {
    // Veille poste partagé : cookies purgés mais overlay de re-auth sur place.
    if (isDeskLockedFlag()) return
    return navigateTo('/login')
  }
})
