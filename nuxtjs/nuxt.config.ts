/**
 * CSP calculée au build : Nuxt requiert 'unsafe-inline' (scripts d'hydratation),
 * Google Sign-In son script/iframe, et le WS pitch une connexion directe à l'API.
 */
function buildCsp(): string {
  const apiBase = (process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291').replace(/\/$/, '')
  const apiWs = apiBase.replace(/^http/, 'ws')
  return [
    "default-src 'self'",
    "script-src 'self' 'unsafe-inline' https://accounts.google.com",
    "style-src 'self' 'unsafe-inline'",
    // Avatars/photos : BFF, data-URI, blob (aperçus upload) et médias GCS/https.
    "img-src 'self' data: blob: https:",
    "font-src 'self'",
    `connect-src 'self' ${apiBase} ${apiWs} https://accounts.google.com`,
    // Audio des comptes rendus (stream authentifié via API).
    `media-src 'self' blob: ${apiBase}`,
    'frame-src https://accounts.google.com',
    "worker-src 'self' blob:",
    "object-src 'none'",
    "base-uri 'self'",
    "form-action 'self'",
    "frame-ancestors 'none'",
  ].join('; ')
}

export default defineNuxtConfig({
  compatibilityDate: '2026-07-15',
  modules: ['@nuxtjs/i18n'],
  i18n: {
    restructureDir: false,
    locales: [
      { code: 'fr', language: 'fr-FR', file: 'fr.json' },
      { code: 'nl', language: 'nl-NL', file: 'nl.json' },
      { code: 'en', language: 'en-GB', file: 'en.json' },
      { code: 'es', language: 'es-ES', file: 'es.json' },
      { code: 'et', language: 'et-EE', file: 'et.json' },
    ],
    defaultLocale: 'fr',
    strategy: 'no_prefix',
    lazy: true,
    langDir: 'locales',
    detectBrowserLanguage: {
      cookieKey: 'pf_locale',
      useCookie: true,
      fallbackLocale: 'fr',
    },
  },
  // Never force-enable in prod builds (Cloud Run OOM risk with Node 22 + SSR).
  devtools: { enabled: process.env.NODE_ENV !== 'production' },
  css: ['~/assets/css/fonts.css', '~/assets/css/tokens.css', '~/assets/css/main.css'],
  runtimeConfig: {
    apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291',
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291',
      googleClientId: process.env.NUXT_PUBLIC_GOOGLE_CLIENT_ID || '',
    },
  },
  routeRules: {
    '/**': {
      headers: {
        'X-Frame-Options': 'DENY',
        'X-Content-Type-Options': 'nosniff',
        'Referrer-Policy': 'strict-origin-when-cross-origin',
        // Micro autorisé (entraînement pitch) ; caméra / géoloc inutiles côté Pro.
        'Permissions-Policy': 'camera=(), geolocation=(), microphone=(self)',
        // Ignoré en HTTP local ; effectif derrière TLS (Cloud Run).
        'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
        'Content-Security-Policy': buildCsp(),
      },
    },
  },
  app: {
    head: {
      title: 'petsFollow Pro',
      // Polices auto-hébergées via assets/css/fonts.css (RGPD : aucun appel Google Fonts).
      link: [
        { rel: 'icon', href: '/brand/emblem.svg' },
        {
          rel: 'preload',
          href: '/fonts/dm-sans-latin.woff2',
          as: 'font',
          type: 'font/woff2',
          crossorigin: 'anonymous',
        },
      ],
      script: [{ src: '/pf-theme-init.js', tagPosition: 'head' }],
    },
  },
})
