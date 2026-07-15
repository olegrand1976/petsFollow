export function useApiError() {
  const { t } = useI18n()

  function translateKey(key: string): string | null {
    const fullKey = key.startsWith('errors.') ? key : `errors.${key}`
    const translated = t(fullKey)
    return translated !== fullKey ? translated : null
  }

  function mapError(e: any): string {
    const apiMessage = e?.data?.error?.message || e?.data?.message
    const msgKey = e?.data?.error?.msgKey || e?.data?.error?.messageKey
    if (msgKey) {
      const translated = translateKey(msgKey)
      if (translated) return translated
    }
    if (apiMessage) return apiMessage
    const code = e?.data?.error?.code
    if (code) {
      const translated = translateKey(code)
      if (translated) return translated
    }
    return t('errors.generic')
  }

  return { mapError }
}
