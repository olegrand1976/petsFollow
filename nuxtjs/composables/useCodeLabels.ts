/** Translate API enum/status codes for admin & commercial tables. */
export function useCodeLabels() {
  const { t, te } = useI18n()

  function roleLabel(role: string): string {
    const key = `common.codes.role.${role}`
    return te(key) ? t(key) : role
  }

  function paymentLabel(code: string): string {
    const key = `common.codes.payment.${code}`
    return te(key) ? t(key) : code
  }

  function planLabel(code: string): string {
    const key = `common.codes.plan.${code}`
    return te(key) ? t(key) : code
  }

  function billingModeLabel(code: string): string {
    const key = `common.codes.billingMode.${code}`
    return te(key) ? t(key) : code
  }

  function prospectStatusLabel(status: string): string {
    const key = `commercial.prospects.status.${status}`
    return te(key) ? t(key) : status
  }

  function prospectSourceLabel(source: string): string {
    const key = `commercial.prospects.source.${source || 'commercial'}`
    return te(key) ? t(key) : source
  }

  return {
    roleLabel,
    paymentLabel,
    planLabel,
    billingModeLabel,
    prospectStatusLabel,
    prospectSourceLabel,
  }
}
