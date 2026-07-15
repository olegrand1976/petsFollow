export default defineNuxtConfig({
  compatibilityDate: '2026-07-15',
  modules: ['@nuxtjs/i18n'],
  i18n: {
    restructureDir: false,
    locales: [
      { code: 'fr', language: 'fr-FR', file: 'fr.json' },
      { code: 'nl', language: 'nl-NL', file: 'nl.json' },
      { code: 'en', language: 'en-GB', file: 'en.json' },
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
  devtools: { enabled: true },
  css: ['~/assets/css/tokens.css', '~/assets/css/main.css'],
  runtimeConfig: {
    apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291',
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291',
      googleClientId: process.env.NUXT_PUBLIC_GOOGLE_CLIENT_ID || '',
    },
  },
  app: {
    head: {
      title: 'petsFollow Pro',
      link: [
        { rel: 'icon', href: '/brand/emblem.svg' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;600;700&family=IBM+Plex+Mono:wght@400;600&display=swap' },
      ],
      script: [{ src: '/pf-theme-init.js', tagPosition: 'head' }],
    },
  },
})
