/** Maps API visit-report payload to editable body + history fields. */
export type VisitReportFields = {
  bodyText: string
  transcriptText: string
  improvedText: string
  status: string
}

export function mapVisitReportFields(
  data: Record<string, unknown> | null | undefined,
): VisitReportFields {
  if (!data) {
    return { bodyText: '', transcriptText: '', improvedText: '', status: '' }
  }
  const transcriptText = String(data.transcriptText ?? '')
  const improvedText = String(data.improvedText ?? '')
  const bodyRaw = String(data.bodyText ?? '')
  const bodyText = bodyRaw || transcriptText
  const status = String(data.status ?? '')
  return { bodyText, transcriptText, improvedText, status }
}

/** Persisted body shown in history only when distinct from transcript / IA. */
export function persistedHistoryBody(
  persistedBody: string,
  transcriptText: string,
  improvedText: string,
): string {
  const body = persistedBody.trim()
  if (!body) return ''
  if (body === transcriptText.trim() || body === improvedText.trim()) return ''
  return persistedBody
}
