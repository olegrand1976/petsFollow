import type { H3Event } from 'h3'

/** Aligné sur JWT_REFRESH_TTL (7 jours). */
const AUTH_COOKIE_MAX_AGE = 7 * 24 * 60 * 60

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

function authCookieOpts() {
  return {
    maxAge: AUTH_COOKIE_MAX_AGE,
    path: '/',
    sameSite: 'lax' as const,
    secure: process.env.NODE_ENV === 'production',
    // Anti-XSS : les JWT ne sont jamais lisibles par le JS navigateur.
    httpOnly: true,
  }
}

/** Marqueur de session non-httpOnly (aucune donnée sensible) pour les middlewares côté client. */
function sessionMarkerOpts() {
  return {
    maxAge: AUTH_COOKIE_MAX_AGE,
    path: '/',
    sameSite: 'lax' as const,
    secure: process.env.NODE_ENV === 'production',
  }
}

export function setAuthCookies(
  event: H3Event,
  pair: { accessToken: string, refreshToken?: string },
) {
  const opts = authCookieOpts()
  setCookie(event, 'pf_token', pair.accessToken, opts)
  if (pair.refreshToken) {
    setCookie(event, 'pf_refresh', pair.refreshToken, opts)
  }
  setCookie(event, 'pf_session', '1', sessionMarkerOpts())
}

export function clearAuthCookies(event: H3Event) {
  const opts = { path: '/', sameSite: 'lax' as const, secure: process.env.NODE_ENV === 'production' }
  deleteCookie(event, 'pf_token', opts)
  deleteCookie(event, 'pf_refresh', opts)
  deleteCookie(event, 'pf_session', opts)
}

/**
 * Absorbe une réponse auth Go : pose les cookies httpOnly et retire les JWT du body
 * renvoyé au navigateur. Les challenges MFA (sans accessToken) passent inchangés.
 */
export function absorbAuthTokens<T>(event: H3Event, res: T): T {
  const envelope = res as { data?: Record<string, unknown> } & Record<string, unknown>
  const data = (envelope?.data ?? envelope) as Record<string, unknown>
  const accessToken = data?.accessToken as string | undefined
  if (!accessToken) return res
  setAuthCookies(event, { accessToken, refreshToken: data.refreshToken as string | undefined })
  const { accessToken: _a, refreshToken: _r, ...rest } = data
  const sanitized = { ...rest, authenticated: true }
  return (envelope?.data ? { ...envelope, data: sanitized } : sanitized) as T
}

type TokenPair = { accessToken: string, refreshToken?: string, expiresIn?: number }

export async function refreshAccessToken(event: H3Event): Promise<TokenPair | null> {
  const refreshToken = getCookie(event, 'pf_refresh')
  if (!refreshToken) return null
  try {
    const res = await $fetch<{ data?: TokenPair } & TokenPair>(`${apiBase()}/api/v1/auth/refresh`, {
      method: 'POST',
      body: { refreshToken },
      headers: localeHeaders(event),
    })
    const pair = (res as { data?: TokenPair }).data ?? (res as TokenPair)
    if (!pair?.accessToken) return null
    setAuthCookies(event, pair)
    return pair
  } catch {
    clearAuthCookies(event)
    return null
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
    const pair = await refreshAccessToken(event)
    if (!pair) throw toProxyError(e)
    try {
      return await fetchOnce(pair.accessToken)
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
    const pair = await refreshAccessToken(event)
    if (!pair) throw toProxyError(e)
    try {
      return await fetchOnce(pair.accessToken)
    } catch (retryErr: any) {
      throw toProxyError(retryErr)
    }
  }
}
