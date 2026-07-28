import { buildCsp } from './utils/buildCsp'

export default defineNuxtConfig({
  compatibilityDate: '2026-07-15',
  modules: ['@nuxtjs/i18n'],
  i18n: {
    restructureDir: false,
    locales: [
      { code: 'fr', language: 'fr-FR', files: ['fr.json', 'pitch-deck/fr.json'] },
      { code: 'nl', language: 'nl-NL', files: ['nl.json', 'pitch-deck/nl.json'] },
      { code: 'en', language: 'en-GB', files: ['en.json', 'pitch-deck/en.json'] },
      { code: 'es', language: 'es-ES', files: ['es.json', 'pitch-deck/es.json'] },
      { code: 'et', language: 'et-EE', files: ['et.json', 'pitch-deck/et.json'] },
      { code: 'it', language: 'it-IT', files: ['it.json', 'pitch-deck/it.json'] },
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
      /** local | staging | production — pages use cases + badge S (staging AUTH). */
      appEnv: process.env.NUXT_PUBLIC_APP_ENV || 'local',
      /** Pharmacie cabinet (CNK / stock / DAF) — mirror PHARMACY_ENABLED. */
      pharmacyEnabled: process.env.NUXT_PUBLIC_PHARMACY_ENABLED === 'true' || process.env.NUXT_PUBLIC_PHARMACY_ENABLED === '1',
      /** Facturation Billit — mirror BILLIT_ENABLED (opt-in). */
      billitEnabled: process.env.NUXT_PUBLIC_BILLIT_ENABLED === 'true' || process.env.NUXT_PUBLIC_BILLIT_ENABLED === '1',
      /** Ordonnances (brouillons + preview PDF) — mirror PRESCRIPTIONS_ENABLED. */
      prescriptionsEnabled: process.env.NUXT_PUBLIC_PRESCRIPTIONS_ENABLED === 'true' || process.env.NUXT_PUBLIC_PRESCRIPTIONS_ENABLED === '1',
    },
  },
  routeRules: {
    '/admin/usecases': { redirect: '/usecases' },
    '/admin/usecases/**': { redirect: '/usecases' },
    '/commercial/usecases': { redirect: '/usecases' },
    '/commercial/usecases/**': { redirect: '/usecases' },
    '/commercial-manager/usecases': { redirect: '/usecases' },
    '/commercial-manager/usecases/**': { redirect: '/usecases' },
    // Le token de partage est dans l'URL : sans no-referrer, un clic vers /register
    // ou /produits (même origine) le transmettrait dans l'en-tête Referer.
    '/dossier/**': {
      headers: {
        'Referrer-Policy': 'no-referrer',
        'X-Robots-Tag': 'noindex, nofollow',
      },
    },
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
        {
          rel: 'preload',
          href: '/fonts/material-symbols-outlined.woff2',
          as: 'font',
          type: 'font/woff2',
          crossorigin: 'anonymous',
        },
      ],
      script: [{ src: '/pf-theme-init.js', tagPosition: 'head' }],
    },
  },
})
