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
      case 'et':
        return 'et-EE'
      case 'it':
        return 'it-IT'
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

  function formatDay(value: string | Date) {
    return new Date(value).toLocaleDateString(dateLocale(), {
      weekday: 'long',
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })
  }

  function formatTime(value: string | Date) {
    return new Date(value).toLocaleTimeString(dateLocale(), {
      hour: '2-digit',
      minute: '2-digit',
    })
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

  return {
    formatDate,
    formatDay,
    formatTime,
    formatCurrency,
    compareStrings,
    dateLocale,
    currencyLocale,
  }
}
