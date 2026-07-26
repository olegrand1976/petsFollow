/**
 * CSP calculée au build : Nuxt requiert 'unsafe-inline' (scripts d'hydratation),
 * Google Sign-In son script/iframe, et le WS pitch une connexion directe à l'API.
 *
 * Ne jamais bake localhost/127.0.0.1 dans la CSP (Docker sans ARG ne doit pas
 * exposer les restes de dev). L'API publique passe via NUXT_PUBLIC_API_BASE.
 */
export function buildCsp(apiBaseEnv = process.env.NUXT_PUBLIC_API_BASE): string {
  const raw = (apiBaseEnv || '').trim().replace(/\/$/, '')
  const isLoopback = !raw || /^https?:\/\/(localhost|127\.0\.0\.1)(:\d+)?$/i.test(raw)
  const apiBase = isLoopback ? '' : raw
  const apiWs = apiBase ? apiBase.replace(/^http/, 'ws') : ''
  const connectApi = apiBase ? ` ${apiBase} ${apiWs}` : ''
  const mediaApi = apiBase ? ` ${apiBase}` : ''
  // Google Identity Services : https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid#content_security_policy
  return [
    "default-src 'self'",
    "script-src 'self' 'unsafe-inline' https://accounts.google.com/gsi/client",
    "style-src 'self' 'unsafe-inline' https://accounts.google.com/gsi/style",
    // Avatars/photos : BFF, data-URI, blob (aperçus upload) et médias GCS/https.
    "img-src 'self' data: blob: https:",
    "font-src 'self'",
    `connect-src 'self'${connectApi} https://accounts.google.com/gsi/`,
    // Audio des comptes rendus (stream authentifié via API).
    `media-src 'self' blob:${mediaApi}`,
    'frame-src https://accounts.google.com/gsi/',
    "worker-src 'self' blob:",
    "object-src 'none'",
    "base-uri 'self'",
    "form-action 'self'",
    "frame-ancestors 'none'",
  ].join('; ')
}
