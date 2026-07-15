const SUPPORTED_LOCALES = ['fr', 'nl', 'en'] as const
export type AppLocale = (typeof SUPPORTED_LOCALES)[number]

export function useLocaleSync() {
  const { locale, setLocale } = useI18n()
  const localeCookie = useCookie('pf_locale')

  async function syncFromUser() {
    try {
      const res: any = await $fetch('/api/me')
      const data = res.data ?? res
      const preferred = data.preferredLocale as string | undefined
      if (preferred && SUPPORTED_LOCALES.includes(preferred as AppLocale)) {
        await setLocale(preferred)
        localeCookie.value = preferred
      }
    } catch {
      /* ignore — user may not be authenticated */
    }
  }

  async function saveLocale(newLocale: AppLocale) {
    await $fetch('/api/me/locale', { method: 'PATCH', body: { locale: newLocale } })
    await setLocale(newLocale)
    localeCookie.value = newLocale
    await useProUser().fetchUser(true)
  }

  return { syncFromUser, saveLocale, locale, supportedLocales: SUPPORTED_LOCALES }
}
