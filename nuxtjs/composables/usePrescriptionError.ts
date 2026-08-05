export function usePrescriptionError() {
  const { t } = useI18n()
  const { mapError: mapApiError } = useApiError()

  function mapError(e: any): string {
    const raw = e?.data?.data?.error ?? e?.data?.error
    const code = raw?.code || e?.data?.code
    switch (code) {
      case 'prescriptions_disabled':
        return t('prescriptions.errorDisabled')
      case 'forbidden':
        return t('prescriptions.errorForbidden')
      case 'not_found':
        return t('prescriptions.errorNotFound')
      case 'invalid_medications':
        return t('prescriptions.errorMedications')
      case 'invalid_format':
        return t('prescriptions.errorFormat')
      case 'prescription_not_draft':
        return t('prescriptions.errorNotDraft')
      case 'payload_too_large':
        return t('prescriptions.errorPayload')
      case 'no_visit_report':
        return t('prescriptions.errorNoVisitReport')
      case 'visit_mismatch':
        return t('prescriptions.errorVisitMismatch')
      case 'not_configured':
      case 'gemini_error':
        return t('prescriptions.errorAi')
      case 'rate_limited':
        return t('prescriptions.errorRateLimited')
      case 'pdf_failed':
      case 'missing_id':
        return t('prescriptions.errorPdf')
      case 'validation_error':
      case 'invalid_valid_until':
      case 'invalid_status':
        return t('prescriptions.errorGeneric')
      default:
        return mapApiError(e) || t('prescriptions.errorGeneric')
    }
  }

  return { mapError }
}
