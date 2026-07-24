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
    return payload.role as string
  } catch {
    return null
  }
}

/** Force de vente : commercial + responsable commercial (profil étendu). */
export function isSalesForceRole(role: string | null | undefined): boolean {
  return role === 'commercial' || role === 'commercial_manager'
}

/** Rôles autorisés sur la face Pro (Nuxt). */
export function isProRole(role: string | null | undefined): boolean {
  return role === 'admin' || role === 'vet' || isSalesForceRole(role)
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

/**
 * Session présente ? Les JWT sont httpOnly : côté client seul le marqueur
 * `pf_session` est visible ; côté SSR les cookies de requête restent lisibles.
 */
export function hasSessionCookie(): boolean {
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
  const opts = authCookieOpts()
  // Efficace en SSR ; côté client les cookies httpOnly ne sont supprimables que par la BFF.
  useCookie('pf_token', opts).value = null
  useCookie('pf_refresh', opts).value = null
  useCookie('pf_session', opts).value = null
  // Avoid stale Pro profile after logout / non-Pro reject / re-login.
  useState('pro-user').value = null
  if (import.meta.client) {
    try {
      await $fetch('/api/auth/logout', { method: 'POST' })
    } catch {
      // best effort — le marqueur pf_session est déjà purgé.
    }
  }
}
