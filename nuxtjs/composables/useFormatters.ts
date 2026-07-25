export function useFormatters() {
  const { locale } = useI18n()

  function dateLocale(): string {
    switch (locale.value) {
      case 'nl':
        return 'nl-NL'
      case 'en':
        return 'en-GB'
      case 'es':
        return 'es-ES'
      default:
        return 'fr-FR'
    }
  }

  function currencyLocale(): string {
    return dateLocale()
  }

  function formatDate(value: string | Date) {
    return new Date(value).toLocaleString(dateLocale())
  }

  function formatCurrency(cents: number) {
    return new Intl.NumberFormat(currencyLocale(), {
      style: 'currency',
      currency: 'EUR',
    }).format(cents / 100)
  }

  function compareStrings(a: string, b: string) {
    return a.localeCompare(b, dateLocale())
  }

  return { formatDate, formatCurrency, compareStrings, dateLocale, currencyLocale }
}
