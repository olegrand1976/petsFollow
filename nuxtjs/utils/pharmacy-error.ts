/** Resolve API pharmacy error codes to i18n (never show raw snake_case codes). */
export function pharmacyErrorMessage(
  t: (key: string) => string,
  e: { data?: { error?: { code?: string, message?: string } }, message?: string } | null | undefined,
  fallbackKey: string,
): string {
  const code = e?.data?.error?.code?.trim()
  if (code) {
    const key = `pharmacy.errors.${code}`
    const msg = t(key)
    if (msg && msg !== key) return msg
  }
  return t(fallbackKey)
}
