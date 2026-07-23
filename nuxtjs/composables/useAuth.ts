/** Aligné sur JWT_REFRESH_TTL (7 jours) — durée cookie ≠ durée JWT access. */
export const AUTH_COOKIE_MAX_AGE = 7 * 24 * 60 * 60

export type AuthTokens = {
  accessToken: string
  refreshToken?: string
  expiresIn?: number
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

function authCookieOpts() {
  return {
    maxAge: AUTH_COOKIE_MAX_AGE,
    sameSite: 'lax' as const,
    // Align with server/utils/api.ts (NODE_ENV === 'production').
    secure: process.env.NODE_ENV === 'production',
    path: '/',
  }
}

export function persistAuthTokens(pair: AuthTokens) {
  const opts = authCookieOpts()
  const access = useCookie('pf_token', opts)
  const refresh = useCookie('pf_refresh', opts)
  access.value = pair.accessToken
  if (pair.refreshToken) {
    refresh.value = pair.refreshToken
  }
}

export function clearAuthTokens() {
  const access = useCookie('pf_token')
  const refresh = useCookie('pf_refresh')
  access.value = null
  refresh.value = null
  // Avoid stale Pro profile after logout / non-Pro reject / re-login.
  useState('pro-user').value = null
}
