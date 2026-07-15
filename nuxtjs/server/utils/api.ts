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
