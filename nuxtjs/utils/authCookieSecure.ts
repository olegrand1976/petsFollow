/**
 * Flag `Secure` des cookies auth (pf_token / pf_refresh / pf_session).
 * Jamais Secure sur HTTP plain (nuxt preview CI, local) — sinon login Playwright casse.
 *
 * Override : AUTH_COOKIE_SECURE=0|1|false|true
 */
export function authCookieSecure(opts?: { protocol?: string }): boolean {
  const o = (process.env.AUTH_COOKIE_SECURE || '').toLowerCase()
  if (o === '0' || o === 'false') return false
  if (o === '1' || o === 'true') return true

  const proto = opts?.protocol
  if (proto === 'http:') return false
  if (proto === 'https:') return true

  const site = (process.env.PETSFOLLOW_PUBLIC_SITE_URL || '').trim()
  if (site.startsWith('http://')) return false
  if (site.startsWith('https://')) return true

  return process.env.NODE_ENV === 'production'
}
