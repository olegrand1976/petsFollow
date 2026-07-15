export function apiBase() {
  const config = useRuntimeConfig()
  return config.apiBase as string
}

export function authHeaders(event: any) {
  const token = getCookie(event, 'pf_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export function localeHeaders(event: any) {
  const locale = getCookie(event, 'pf_locale')
  return locale ? { 'Accept-Language': locale } : {}
}

export function apiHeaders(event: any) {
  return { ...authHeaders(event), ...localeHeaders(event) }
}

export async function proxyApi<T>(
  event: any,
  path: string,
  options: { method?: string, body?: unknown } = {},
): Promise<T> {
  try {
    return await $fetch<T>(`${apiBase()}${path}`, {
      method: options.method,
      body: options.body,
      headers: apiHeaders(event),
    })
  } catch (e: any) {
    throw createError({
      statusCode: e?.statusCode ?? e?.status ?? 500,
      statusMessage: e?.data?.error?.message ?? e?.statusMessage ?? 'Error',
      data: e?.data,
    })
  }
}
