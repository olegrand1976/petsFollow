/**
 * Locales de la face Pro. Doit rester le miroir exact du tableau `i18n.locales`
 * de `nuxt.config.ts` : une locale listée ici mais absente de la config ferait
 * échouer `setLocale()`, et l'inverse la rendrait injoignable.
 *
 * L'API en supporte deux de plus — `uk` et `ru` (cf. `i18n.Supported` côté Go,
 * utilisées par les e-mails, SMS et push clients Flutter). Elles ne reviennent
 * ici qu'avec un catalogue Nuxt complet : `locales/{uk,ru}.json`,
 * `presentation/` et `ai-flows/`. Tant qu'elles sont absentes,
 * `applyPreferredLocale` ignore une préférence `uk`/`ru` et la face Pro reste
 * en français — c'est volontaire, mieux qu'une UI à moitié traduite.
 */
const SUPPORTED_LOCALES = ['fr', 'nl', 'en', 'es', 'et', 'it'] as const
export type AppLocale = (typeof SUPPORTED_LOCALES)[number]

export function useLocaleSync() {
  const { locale, setLocale } = useI18n()
  const localeCookie = useCookie('pf_locale')

  async function applyPreferredLocale(preferred?: string | null) {
    if (preferred && SUPPORTED_LOCALES.includes(preferred as AppLocale)) {
      await setLocale(preferred)
      localeCookie.value = preferred
    }
  }

  async function syncFromUser() {
    try {
      const res: any = await $fetch('/api/me')
      const data = res.data ?? res
      await applyPreferredLocale(data.preferredLocale as string | undefined)
    } catch {
      /* ignore — user may not be authenticated */
    }
  }

  async function saveLocale(newLocale: AppLocale) {
    await $fetch('/api/me/locale', { method: 'PATCH', body: { locale: newLocale } })
    await setLocale(newLocale)
    localeCookie.value = newLocale
    await useProUser().fetchUser(true).catch(() => {})
  }

  async function switchLocale(newLocale: AppLocale) {
    await setLocale(newLocale)
    localeCookie.value = newLocale
  }

  return {
    syncFromUser,
    applyPreferredLocale,
    saveLocale,
    switchLocale,
    locale,
    supportedLocales: SUPPORTED_LOCALES,
  }
}
