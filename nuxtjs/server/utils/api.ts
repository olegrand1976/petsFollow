import type { H3Event } from 'h3'
import { createError } from 'h3'
import { authCookieSecure } from '../../utils/authCookieSecure'

/** Aligné sur JWT_REFRESH_TTL (30 jours). */
const AUTH_COOKIE_MAX_AGE = 30 * 24 * 60 * 60

export function apiBase() {
  const config = useRuntimeConfig()
  return config.apiBase as string
}

export function authHeaders(event: H3Event) {
  const token = getCookie(event, 'pf_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export function localeHeaders(event: H3Event) {
  const locale = getCookie(event, 'pf_locale')
  const headers: Record<string, string> = locale ? { 'Accept-Language': locale } : {}
  // Rate limit Go par IP réelle : sans ce header, tout le trafic web partagerait l'IP de la BFF.
  const ip = getRequestIP(event, { xForwardedFor: true })
  if (ip) headers['X-Forwarded-For'] = ip
  return headers
}

export function apiHeaders(event: H3Event) {
  return { ...authHeaders(event), ...localeHeaders(event) }
}

function requestProtocol(event: H3Event): string | undefined {
  try {
    return getRequestURL(event).protocol
  } catch {
    return undefined
  }
}

function authCookieOpts(event: H3Event) {
  return {
    maxAge: AUTH_COOKIE_MAX_AGE,
    path: '/',
    sameSite: 'lax' as const,
    secure: authCookieSecure({ protocol: requestProtocol(event) }),
    // Anti-XSS : les JWT ne sont jamais lisibles par le JS navigateur.
    httpOnly: true,
  }
}

/** Marqueur de session non-httpOnly (aucune donnée sensible) pour les middlewares côté client. */
function sessionMarkerOpts(event: H3Event) {
  return {
    maxAge: AUTH_COOKIE_MAX_AGE,
    path: '/',
    sameSite: 'lax' as const,
    secure: authCookieSecure({ protocol: requestProtocol(event) }),
  }
}

export function setAuthCookies(
  event: H3Event,
  pair: { accessToken: string, refreshToken?: string },
) {
  const opts = authCookieOpts(event)
  setCookie(event, 'pf_token', pair.accessToken, opts)
  if (pair.refreshToken) {
    setCookie(event, 'pf_refresh', pair.refreshToken, opts)
  }
  setCookie(event, 'pf_session', '1', sessionMarkerOpts(event))
}

export function clearAuthCookies(event: H3Event) {
  // httpOnly must match setAuthCookies or browsers may keep pf_token / pf_refresh.
  const secure = authCookieSecure({ protocol: requestProtocol(event) })
  const tokenOpts = {
    path: '/',
    sameSite: 'lax' as const,
    secure,
    httpOnly: true,
  }
  const markerOpts = {
    path: '/',
    sameSite: 'lax' as const,
    secure,
  }
  deleteCookie(event, 'pf_token', tokenOpts)
  deleteCookie(event, 'pf_refresh', tokenOpts)
  deleteCookie(event, 'pf_session', markerOpts)
}

/**
 * Purger les cookies ne suffit pas : un refresh token qui aurait fuité resterait
 * valide 30 jours. L'API incrémente token_version pour le rendre inutilisable.
 * Un échec ne doit pas empêcher la déconnexion locale.
 */
export async function revokeIssuedTokens(event: H3Event) {
  try {
    await $fetch(`${apiBase()}/api/v1/auth/logout`, {
      method: 'POST',
      headers: { ...localeHeaders(event), ...authHeaders(event) },
    })
  }
  catch {
    // session déjà expirée ou API indisponible : on purge quand même les cookies
  }
}

/**
 * Absorbe une réponse auth Go : pose les cookies httpOnly et retire les JWT du body
 * renvoyé au navigateur. Les challenges MFA (sans accessToken) passent inchangés.
 * Expose `role` (claim JWT) pour la navigation document post-login sans XHR /me.
 */
export function absorbAuthTokens<T>(event: H3Event, res: T): T {
  const envelope = res as { data?: Record<string, unknown> } & Record<string, unknown>
  const data = (envelope?.data ?? envelope) as Record<string, unknown>
  const accessToken = data?.accessToken as string | undefined
  if (!accessToken) return res
  setAuthCookies(event, { accessToken, refreshToken: data.refreshToken as string | undefined })
  const { accessToken: _a, refreshToken: _r, ...rest } = data
  const role = roleFromAccessToken(accessToken)
  const sanitized = {
    ...rest,
    authenticated: true,
    ...(role ? { role } : {}),
  }
  return (envelope?.data ? { ...envelope, data: sanitized } : sanitized) as T
}

/** Claim `role` du JWT access — pas de vérif crypto (token déjà émis par notre API). */
export function roleFromAccessToken(accessToken: string): string | undefined {
  const parts = accessToken.split('.')
  if (parts.length < 2) return undefined
  try {
    const b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const json = Buffer.from(b64, 'base64').toString('utf8')
    const payload = JSON.parse(json) as { role?: unknown }
    return typeof payload.role === 'string' ? payload.role : undefined
  } catch {
    return undefined
  }
}

type TokenPair = { accessToken: string, refreshToken?: string, expiresIn?: number }

/** Résultat refresh — permet au proxy de renvoyer 503 (pas 401) si l'API est momentanément down. */
export type RefreshOutcome =
  | { kind: 'ok', pair: TokenPair }
  | { kind: 'missing' }
  | { kind: 'rejected' }
  | { kind: 'transient', status?: number }

/** 401/403 = session morte ; tout le reste (5xx, réseau) = garder les cookies. */
export function classifyRefreshHttpStatus(status: number | undefined): 'rejected' | 'transient' {
  if (status === 401 || status === 403) return 'rejected'
  return 'transient'
}

export async function refreshAccessToken(event: H3Event): Promise<RefreshOutcome> {
  const refreshToken = getCookie(event, 'pf_refresh')
  if (!refreshToken) return { kind: 'missing' }
  try {
    const res = await $fetch<{ data?: TokenPair } & TokenPair>(`${apiBase()}/api/v1/auth/refresh`, {
      method: 'POST',
      body: { refreshToken },
      headers: localeHeaders(event),
    })
    const pair = (res as { data?: TokenPair }).data ?? (res as TokenPair)
    if (!pair?.accessToken) {
      console.warn('[auth/refresh] rejected, empty accessToken')
      clearAuthCookies(event)
      return { kind: 'rejected' }
    }
    setAuthCookies(event, pair)
    return { kind: 'ok', pair }
  } catch (e: unknown) {
    const err = e as { statusCode?: number, status?: number }
    const status = err?.statusCode ?? err?.status
    const kind = classifyRefreshHttpStatus(status)
    if (kind === 'rejected') {
      console.warn('[auth/refresh] rejected, clearing cookies', { status })
      clearAuthCookies(event)
      return { kind: 'rejected' }
    }
    console.warn('[auth/refresh] transient failure, keeping cookies', { status: status ?? 'unknown' })
    return { kind: 'transient', status }
  }
}

function refreshUnavailableError() {
  return createError({
    statusCode: 503,
    statusMessage: 'Auth refresh temporarily unavailable',
    data: { reason: 'refresh_unavailable' },
  })
}

/** Après un 401 upstream : retry si refresh OK, sinon 401 session morte ou 503 transient. */
function throwAfterFailedRefresh(outcome: RefreshOutcome, originalUnauthorized: unknown): never {
  switch (outcome.kind) {
    case 'ok':
      // Invariant : l'appelant ne passe ici que si refresh a échoué.
      throw toProxyError(originalUnauthorized)
    case 'transient':
      throw refreshUnavailableError()
    case 'missing':
    case 'rejected':
      throw toProxyError(originalUnauthorized)
    default: {
      const _exhaustive: never = outcome
      throw _exhaustive
    }
  }
}

function isUnauthorized(e: any): boolean {
  const status = e?.statusCode ?? e?.status
  return status === 401
}

function toProxyError(e: any) {
  return createError({
    statusCode: e?.statusCode ?? e?.status ?? 500,
    statusMessage: e?.data?.error?.message ?? e?.statusMessage ?? 'Error',
    data: e?.data,
  })
}

/** Proxy sans retry refresh — pour routes auth publiques (login, register, etc.). */
export async function proxyPublicApi<T>(
  event: H3Event,
  path: string,
  options: { method?: string, body?: unknown, headers?: Record<string, string> } = {},
): Promise<T> {
  try {
    return await $fetch<T>(`${apiBase()}${path}`, {
      method: options.method,
      body: options.body,
      headers: { ...localeHeaders(event), ...options.headers },
    })
  } catch (e: any) {
    throw toProxyError(e)
  }
}

function bearerHeaders(event: H3Event, accessToken: string, extra?: Record<string, string>) {
  return {
    ...localeHeaders(event),
    Authorization: `Bearer ${accessToken}`,
    ...extra,
  }
}

export async function proxyApi<T>(
  event: H3Event,
  path: string,
  options: {
    method?: string
    body?: unknown
    headers?: Record<string, string>
    query?: Record<string, unknown>
  } = {},
): Promise<T> {
  const url = `${apiBase()}${path}`
  const fetchOnce = (accessToken?: string) =>
    $fetch<T>(url, {
      method: options.method,
      body: options.body,
      query: options.query,
      headers: accessToken
        ? bearerHeaders(event, accessToken, options.headers)
        : { ...apiHeaders(event), ...options.headers },
    })

  try {
    return await fetchOnce()
  } catch (e: any) {
    if (!isUnauthorized(e)) throw toProxyError(e)
    const outcome = await refreshAccessToken(event)
    if (outcome.kind !== 'ok') throwAfterFailedRefresh(outcome, e)
    try {
      return await fetchOnce(outcome.pair.accessToken)
    } catch (retryErr: any) {
      throw toProxyError(retryErr)
    }
  }
}

/** Proxy multipart/binary uploads with one refresh+retry on 401. */
export async function proxyUpload(
  event: H3Event,
  path: string,
  body: Buffer | Uint8Array | string | undefined,
  contentType: string,
) {
  const token = getCookie(event, 'pf_token')
  const refresh = getCookie(event, 'pf_refresh')
  if (!token && !refresh) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  const url = `${apiBase()}${path}`
  const fetchOnce = (accessToken?: string) =>
    $fetch(url, {
      method: 'POST',
      body,
      headers: accessToken
        ? bearerHeaders(event, accessToken, { 'content-type': contentType })
        : { ...apiHeaders(event), 'content-type': contentType },
    })

  try {
    return await fetchOnce()
  } catch (e: any) {
    if (!isUnauthorized(e)) throw toProxyError(e)
    const outcome = await refreshAccessToken(event)
    if (outcome.kind !== 'ok') throwAfterFailedRefresh(outcome, e)
    try {
      return await fetchOnce(outcome.pair.accessToken)
    } catch (retryErr: any) {
      throw toProxyError(retryErr)
    }
  }
}

/** Proxy binary GET (DICOM / preview / PDF) — returns a Node Buffer to the client. */
export async function proxyBinary(
  event: H3Event,
  path: string,
  opts?: { contentDispositionFallback?: string },
) {
  const url = `${apiBase()}${path}`
  const fetchOnce = async (accessToken?: string) => {
    const headers = accessToken
      ? bearerHeaders(event, accessToken)
      : { ...apiHeaders(event) }
    const res = await fetch(url, { headers })
    if (!res.ok) {
      let payload: any = null
      try {
        payload = await res.json()
      } catch { /* non-JSON upstream */ }
      const err: any = createError({
        statusCode: res.status,
        statusMessage: payload?.error?.message || payload?.message || res.statusText || 'Error',
        data: payload?.error || payload || { code: 'upstream_error', message: `upstream_${res.status}` },
      })
      throw err
    }
    // Buffer (not ArrayBuffer): Nitro/H3 can stall when serializing raw ArrayBuffer.
    const buf = Buffer.from(await res.arrayBuffer())
    const ct = res.headers.get('content-type') || 'application/octet-stream'
    setHeader(event, 'content-type', ct)
    setHeader(event, 'cache-control', res.headers.get('cache-control') || 'private, no-store')
    setHeader(event, 'content-length', String(buf.length))
    const cd = res.headers.get('content-disposition') || opts?.contentDispositionFallback
    if (cd) setHeader(event, 'content-disposition', cd)
    return buf
  }

  try {
    return await fetchOnce()
  } catch (e: any) {
    if (!isUnauthorized(e)) {
      if (e?.statusCode) throw e
      throw toProxyError(e)
    }
    const outcome = await refreshAccessToken(event)
    if (outcome.kind !== 'ok') throwAfterFailedRefresh(outcome, e)
    try {
      return await fetchOnce(outcome.pair.accessToken)
    } catch (retryErr: any) {
      if (retryErr?.statusCode) throw retryErr
      throw toProxyError(retryErr)
    }
  }
}
