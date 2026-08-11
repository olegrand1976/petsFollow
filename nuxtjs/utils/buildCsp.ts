/**
 * CSP calculée au build : Nuxt requiert 'unsafe-inline' (scripts d'hydratation),
 * Google Sign-In son script/iframe, et le WS pitch une connexion directe à l'API.
 *
 * - Pas de défaut localhost : env absente → `'self'` seul (évite bake Docker nu).
 * - Loopback explicite (CI Playwright `NUXT_PUBLIC_API_BASE=http://localhost:8291`) conservé
 *   pour les WebSockets pitch.
 * - Prod Cloud Build : ARG `https://api.petsfollow.ll-it-sc.be`.
 * - `frameAncestors: "'self'"` : réponses PDF admin (viewer iframe same-origin).
 */
export type BuildCspOpts = {
  frameAncestors?: "'none'" | "'self'"
}

export function buildCsp(
  apiBaseEnv = process.env.NUXT_PUBLIC_API_BASE,
  opts?: BuildCspOpts,
): string {
  const apiBase = (apiBaseEnv || '').trim().replace(/\/$/, '')
  const apiWs = apiBase ? apiBase.replace(/^http/, 'ws') : ''
  const connectApi = apiBase ? ` ${apiBase} ${apiWs}` : ''
  const mediaApi = apiBase ? ` ${apiBase}` : ''
  const frameAncestors = opts?.frameAncestors ?? "'none'"
  // Google Identity Services : https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid#content_security_policy
  return [
    "default-src 'self'",
    "script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval' https://accounts.google.com/gsi/client",
    "style-src 'self' 'unsafe-inline' https://accounts.google.com/gsi/style",
    // Avatars/photos : BFF, data-URI, blob (aperçus upload) et médias GCS/https.
    "img-src 'self' data: blob: https:",
    "font-src 'self'",
    `connect-src 'self'${connectApi} https://accounts.google.com/gsi/`,
    // Audio des comptes rendus (stream authentifié via API).
    `media-src 'self' blob:${mediaApi}`,
    'frame-src \'self\' https://accounts.google.com/gsi/',
    "worker-src 'self' blob:",
    "object-src 'none'",
    "base-uri 'self'",
    "form-action 'self'",
    `frame-ancestors ${frameAncestors}`,
  ].join('; ')
}
