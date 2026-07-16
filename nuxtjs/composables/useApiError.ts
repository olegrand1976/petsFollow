export function useApiError() {
  const { t } = useI18n()

  function translateKey(key: string): string | null {
    const fullKey = key.startsWith('errors.') ? key : `errors.${key}`
    const translated = t(fullKey)
    return translated !== fullKey ? translated : null
  }

  function mapError(e: any): string {
    // Nitro createError: corps Go dans e.data.data ; ofetch direct: e.data
    const raw = e?.data?.data?.error ?? e?.data?.error
    const apiErr = raw && typeof raw === 'object' ? raw : null
    const apiMessage = apiErr?.message || e?.data?.message || e?.statusMessage
    const msgKey = apiErr?.msgKey || apiErr?.messageKey
    if (msgKey) {
      const translated = translateKey(msgKey)
      if (translated) return translated
    }
    if (apiErr?.code) {
      const translated = translateKey(apiErr.code)
      if (translated) return translated
    }
    if (apiMessage && apiMessage !== 'Server Error') return apiMessage
    if (e?.statusCode === 401 || e?.data?.statusCode === 401) {
      const translated = translateKey('unauthorized')
      if (translated) return translated
    }
    return t('errors.generic')
  }

  return { mapError }
}
