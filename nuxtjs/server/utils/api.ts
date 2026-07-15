export function apiBase() {
  const config = useRuntimeConfig()
  return config.apiBase as string
}

export function authHeaders(event: any) {
  const token = getCookie(event, 'pf_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}
