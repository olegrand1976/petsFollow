/** Aligné sur JWT_REFRESH_TTL (7 jours) — durée cookie ≠ durée JWT access. */
export const AUTH_COOKIE_MAX_AGE = 7 * 24 * 60 * 60

export type AuthTokens = {
  /** Absent quand la BFF a absorbé les tokens en cookies httpOnly. */
  accessToken?: string
  refreshToken?: string
  expiresIn?: number
  /** Posé par la BFF quand les cookies httpOnly ont été établis. */
  authenticated?: boolean
}

export type AuthMFAChallenge = {
  requires2FA: true
  mfaToken: string
  expiresIn?: number
}

export type AuthResponse = AuthTokens | AuthMFAChallenge

export function isMFAChallenge(res: AuthResponse): res is AuthMFAChallenge {
  return 'requires2FA' in res && res.requires2FA === true
}

export function parseJwtRole(token: string | null | undefined): string | null {
  if (!token) return null
  const parts = token.split('.')
  if (parts.length < 2) return null
  try {
    const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')))
    // Expired access JWT must not drive AUTH_ENTRY redirects (SSR loop risk).
    if (typeof payload.exp === 'number' && payload.exp * 1000 <= Date.now()) {
      return null
    }
    return (payload.role as string) || null
  } catch {
    return null
  }
}

/** Force de vente : commercial + responsable commercial (profil étendu). */
export function isSalesForceRole(role: string | null | undefined): boolean {
  return role === 'commercial' || role === 'commercial_manager'
}

/** Staff cabinet : véto référence / assistant / secrétaire. */
export function isPracticeStaffRole(role: string | null | undefined): boolean {
  return role === 'vet' || role === 'vet_assistant' || role === 'secretary'
}

/** Rôles autorisés sur la face Pro (Nuxt). */
export function isProRole(role: string | null | undefined): boolean {
  return role === 'admin' || isPracticeStaffRole(role) || isSalesForceRole(role)
}

/** Home post-login / post-change-password pour un rôle Pro. */
export function homePathForRole(role: string | null | undefined, opts?: { profileComplete?: boolean | null }): string {
  switch (role) {
    case 'admin':
      return '/admin'
    case 'commercial_manager':
      return '/commercial-manager'
    case 'commercial':
      return '/commercial'
    case 'vet':
    case 'vet_assistant':
    case 'secretary':
      return opts?.profileComplete === false ? '/onboarding' : '/dashboard'
    default:
      return '/login'
  }
}

export function unwrapAuthData(res: unknown): AuthResponse {
  const data = (res as { data?: AuthResponse })?.data ?? res
  return data as AuthResponse
}

export function extractAccessToken(res: AuthResponse): string | null {
  if (isMFAChallenge(res)) return null
  return res.accessToken ?? null
}

/** Succès auth : cookies httpOnly posés par la BFF (`authenticated`) ou tokens legacy. */
export function isAuthSuccess(res: AuthResponse): boolean {
  if (isMFAChallenge(res)) return false
  return res.authenticated === true || !!res.accessToken
}

function authCookieOpts() {
  return {
    sameSite: 'lax' as const,
    // Align with server/utils/api.ts (NODE_ENV === 'production').
    secure: process.env.NODE_ENV === 'production',
    path: '/',
  }
}

/** Opts de purge JWT — httpOnly doit matcher setAuthCookies BFF. */
function httpOnlyAuthCookieOpts() {
  return {
    ...authCookieOpts(),
    httpOnly: true,
  }
}

/** Marqueur pf_session (non-httpOnly) — aligné sur sessionMarkerOpts BFF. */
export function sessionCookieOpts() {
  return {
    ...authCookieOpts(),
    maxAge: AUTH_COOKIE_MAX_AGE,
  }
}

/**
 * Après clear dans la même requête SSR, getCookie peut encore renvoyer l'ancien
 * JWT (Set-Cookie ne met pas à jour la map request) → boucle login↔dashboard.
 */
function authClearedState() {
  return useState<boolean>('pf-auth-cleared', () => false)
}

/** Appeler après login / confirm réussi (SPA) pour réactiver hasSessionCookie. */
export function markAuthSessionActive() {
  authClearedState().value = false
}

/**
 * Session présente ? Les JWT sont httpOnly : côté client seul le marqueur
 * `pf_session` est visible ; côté SSR les cookies de requête restent lisibles.
 */
export function hasSessionCookie(): boolean {
  if (authClearedState().value) return false
  return !!(
    useCookie('pf_token').value
    || useCookie('pf_refresh').value
    || useCookie('pf_session').value
  )
}

/** Token d'accès (TTL court) pour les WebSockets — les cookies étant httpOnly. */
export async function fetchWsToken(): Promise<string> {
  try {
    const res: any = await $fetch('/api/auth/ws-token')
    const data = res?.data ?? res
    return (data?.token as string) || ''
  } catch {
    return ''
  }
}

export async function clearAuthTokens() {
  authClearedState().value = true
  // Avoid stale Pro profile after logout / non-Pro reject / re-login.
  useState('pro-user').value = null
  // pf_session is the only auth cookie the browser JS can clear (non-httpOnly).
  useCookie('pf_session', sessionCookieOpts()).value = null
  if (import.meta.server) {
    // SSR: Set-Cookie can clear httpOnly JWT cookies on the response.
    useCookie('pf_token', httpOnlyAuthCookieOpts()).value = null
    useCookie('pf_refresh', httpOnlyAuthCookieOpts()).value = null
  }
  if (import.meta.client) {
    try {
      // httpOnly pf_token / pf_refresh: only the BFF may delete them.
      await $fetch('/api/auth/logout', { method: 'POST' })
    } catch {
      // best effort — le marqueur pf_session est déjà purgé.
    }
  }
}
