import { authCookieSecure } from '../utils/authCookieSecure'
import { needsContactPhone } from '../utils/needsContactPhone'

/** Aligné sur JWT_REFRESH_TTL (30 jours) — durée cookie ≠ durée JWT access. */
export const AUTH_COOKIE_MAX_AGE = 30 * 24 * 60 * 60

export type AuthTokens = {
  /** Absent quand la BFF a absorbé les tokens en cookies httpOnly. */
  accessToken?: string
  refreshToken?: string
  expiresIn?: number
  /** Posé par la BFF quand les cookies httpOnly ont été établis. */
  authenticated?: boolean
  /** Claim rôle (BFF, après absorption) — pour redirect document sans XHR /me. */
  role?: string
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

/** Profil minimal lu après login (GET /api/me). */
export type PostLoginProfile = {
  role?: string
  profileComplete?: boolean | null
  mustChangePassword?: boolean | null
  contactPhone?: string | null
  preferredLocale?: string | null
}

/**
 * Décision post-login : ne pas confondre « /me indisponible » (cookies / course WebKit)
 * avec « rôle non-Pro ».
 */
export type PostLoginTarget =
  | { kind: 'navigate'; path: string }
  | { kind: 'proOnly' }
  | { kind: 'sessionUnavailable' }

export function resolvePostLoginTarget(
  me: PostLoginProfile | null | undefined,
  jwtRole?: string | null,
): PostLoginTarget {
  const role = me?.role || jwtRole || null
  if (!role) {
    return { kind: 'sessionUnavailable' }
  }
  if (me?.mustChangePassword === true) {
    return { kind: 'navigate', path: '/change-password' }
  }
  // Only when /me loaded — JWT-only fallback must not assume empty phone.
  if (me && needsContactPhone(me)) {
    return { kind: 'navigate', path: '/complete-contact-phone' }
  }
  if (!isProRole(role)) {
    return { kind: 'proOnly' }
  }
  return {
    kind: 'navigate',
    path: homePathForRole(role, { profileComplete: me?.profileComplete }),
  }
}

export function authErrorStatus(e: unknown): number | null {
  const err = e as { statusCode?: number; status?: number; response?: { status?: number } }
  return err?.statusCode ?? err?.status ?? err?.response?.status ?? null
}

/**
 * Relit /api/me juste après Set-Cookie (course possible sur WebKit / Firefox iOS).
 * Réessaie sur soft-null et 401/403 ; propage le dernier 401/403 si toujours KO.
 */
export async function fetchUserAfterLogin<T extends PostLoginProfile>(
  fetchUser: (force?: boolean) => Promise<T | null>,
  opts?: {
    attempts?: number
    delayMs?: number
    sleep?: (ms: number) => Promise<void>
  },
): Promise<T | null> {
  const attempts = opts?.attempts ?? 3
  const delayMs = opts?.delayMs ?? 120
  const sleep = opts?.sleep ?? ((ms: number) => new Promise((r) => setTimeout(r, ms)))
  let lastAuthError: unknown = null

  for (let i = 0; i < attempts; i++) {
    try {
      const me = await fetchUser(true)
      if (me?.role) return me
    } catch (e) {
      const status = authErrorStatus(e)
      if (status === 401 || status === 403) {
        lastAuthError = e
      } else {
        throw e
      }
    }
    if (i < attempts - 1) await sleep(delayMs)
  }
  if (lastAuthError) throw lastAuthError
  return null
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
  const protocol =
    import.meta.client && typeof location !== 'undefined' ? location.protocol : undefined
  return {
    sameSite: 'lax' as const,
    // Align with server/utils/api.ts (authCookieSecure — pas Secure sur HTTP).
    secure: authCookieSecure({ protocol }),
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
  armPostLoginGrace()
}

/** Fenêtre anti-logout après login (course cookies WebKit / hard navigation). */
export const AUTH_POST_LOGIN_GRACE_MS = 15_000

const POST_LOGIN_GRACE_KEY = 'pf_auth_grace_until'

function armPostLoginGrace() {
  if (typeof sessionStorage === 'undefined') return
  try {
    sessionStorage.setItem(POST_LOGIN_GRACE_KEY, String(Date.now() + AUTH_POST_LOGIN_GRACE_MS))
  } catch {
    /* private mode / quota */
  }
}

/** True pendant quelques secondes après finishClientLoginSession.
 * Client-only (sessionStorage) : survit au reload document, mais pas au SSR —
 * le 1er passage middleware après replace s'appuie sur les cookies httpOnly de la requête.
 */
export function isWithinPostLoginGrace(): boolean {
  if (typeof sessionStorage === 'undefined') return false
  try {
    const until = Number(sessionStorage.getItem(POST_LOGIN_GRACE_KEY) || 0)
    return Number.isFinite(until) && Date.now() < until
  } catch {
    return false
  }
}

/**
 * clearAuthTokens sauf pendant la grâce post-login (évite d'éjecter une session naissante).
 * @returns true si la session a bien été purgée.
 */
export async function clearAuthTokensUnlessPostLoginGrace(): Promise<boolean> {
  if (isWithinPostLoginGrace()) return false
  await clearAuthTokens()
  return true
}

/**
 * Path d'entrée post-login (document navigation).
 * Doit rester dans AUTH_ENTRY_PATHS : le middleware SSR lit les cookies httpOnly
 * et redirige vers la home du rôle (évite GET /api/me XHR avant commit WebKit).
 */
export const AUTH_POST_LOGIN_RELOAD_PATH = '/'

/** Query `?reason=` sur /login après refus d'un rôle non-Pro. */
export const AUTH_LOGIN_REASON_PRO_ONLY = 'proOnly'

function safeInternalPath(path: string): string {
  // Empêche open-redirect si un appelant passe une URL absolue / protocol-relative.
  if (!path.startsWith('/') || path.startsWith('//') || path.includes('://')) {
    return AUTH_POST_LOGIN_RELOAD_PATH
  }
  return path
}

/**
 * Finalise un login réussi côté navigateur : marqueur + reload document.
 * Sur Safari/iPad, les Set-Cookie du POST login ne sont pas encore visibles au
 * prochain XHR → 401 /api/me → clearAuthTokens effaçait la session naissante.
 *
 * Ne pas réécrire pf_session en document.cookie : le BFF l'a déjà posé (Set-Cookie).
 * Une écriture client peut créer une session fantôme (SSR sans JWT httpOnly).
 */
export function finishClientLoginSession(path: string = AUTH_POST_LOGIN_RELOAD_PATH) {
  markAuthSessionActive()
  const target = safeInternalPath(path)
  // typeof window : OK en SPA / tests ; absent en SSR Nitro.
  // replace : évite /login dans l'historique (retour → re-redirect).
  if (typeof window !== 'undefined' && typeof window.location?.replace === 'function') {
    window.location.replace(target)
  }
}

/**
 * Session présente ?
 * - SSR : seuls pf_token / pf_refresh comptent (évite session fantôme pf_session sans JWT
 *   → layout Pro rendu au-dessus de /login, hydratation cassée).
 * - Client : pf_session (seul cookie auth lisible JS) + JWT si exposés (legacy).
 */
export function hasSessionCookie(): boolean {
  if (authClearedState().value) return false
  if (import.meta.server) {
    return !!(useCookie('pf_token').value || useCookie('pf_refresh').value)
  }
  return !!(
    useCookie('pf_session').value
    || useCookie('pf_token').value
    || useCookie('pf_refresh').value
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
