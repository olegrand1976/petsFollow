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
  return locale ? { 'Accept-Language': locale } : {}
}

export function apiHeaders(event: H3Event) {
  return { ...authHeaders(event), ...localeHeaders(event) }
}

export function setAuthCookies(
  event: H3Event,
  pair: { accessToken: string, refreshToken?: string },
) {
  setCookie(event, 'pf_token', pair.accessToken, {
    maxAge: AUTH_COOKIE_MAX_AGE,
    path: '/',
    sameSite: 'lax',
  })
  if (pair.refreshToken) {
    setCookie(event, 'pf_refresh', pair.refreshToken, {
      maxAge: AUTH_COOKIE_MAX_AGE,
      path: '/',
      sameSite: 'lax',
    })
  }
}

export function clearAuthCookies(event: H3Event) {
  deleteCookie(event, 'pf_token', { path: '/' })
  deleteCookie(event, 'pf_refresh', { path: '/' })
}

type TokenPair = { accessToken: string, refreshToken?: string, expiresIn?: number }

async function refreshAccessToken(event: H3Event): Promise<TokenPair | null> {
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
  options: { method?: string, body?: unknown, headers?: Record<string, string> } = {},
): Promise<T> {
  const url = `${apiBase()}${path}`
  const fetchOnce = (accessToken?: string) =>
    $fetch<T>(url, {
      method: options.method,
      body: options.body,
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
