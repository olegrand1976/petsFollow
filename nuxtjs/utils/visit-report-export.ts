/** Client-side export helpers for visit CR (markdown + PDF). */

export async function copyVisitReportMarkdown(text: string): Promise<void> {
  const body = String(text || '')
  if (!body.trim()) {
    const err: any = new Error('report_empty')
    err.data = { code: 'report_empty' }
    throw err
  }
  if (typeof navigator === 'undefined' || !navigator.clipboard?.writeText) {
    const err: any = new Error('clipboard_unavailable')
    err.data = { code: 'clipboard_unavailable' }
    throw err
  }
  await navigator.clipboard.writeText(body)
}

export function downloadVisitReportMarkdown(text: string, filenameBase = 'cr'): void {
  const body = String(text || '')
  if (!body.trim()) {
    const err: any = new Error('report_empty')
    err.data = { code: 'report_empty' }
    throw err
  }
  const safe = String(filenameBase || 'cr').replace(/[^\w.-]+/g, '-').replace(/^-+|-+$/g, '') || 'cr'
  const blob = new Blob([body], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${safe}.md`
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  a.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
}

/** Open / download the visit CR PDF via BFF blob (mirrors prescription-pdf). */
export async function openVisitReportPdfBlob(
  visitId: string,
): Promise<'tab' | 'download'> {
  const id = String(visitId || '').trim()
  if (!id) {
    const err: any = new Error('missing_id')
    err.data = { code: 'missing_id' }
    throw err
  }

  const blob = await $fetch<Blob>(`/api/visits/${id}/report-pdf`, {
    responseType: 'blob',
  })

  const type = (blob.type || '').toLowerCase()
  if (type.includes('json') || type.includes('text/html') || type.includes('text/plain')) {
    let msg = 'pdf_failed'
    try {
      const text = await blob.text()
      const parsed = JSON.parse(text) as {
        error?: { code?: string, message?: string }
        statusMessage?: string
        data?: { code?: string }
      }
      msg = parsed?.error?.code
        || parsed?.data?.code
        || parsed?.error?.message
        || parsed?.statusMessage
        || msg
    }
    catch { /* keep default */ }
    const err: any = new Error(msg)
    err.data = { code: msg }
    throw err
  }

  const typed = type.includes('pdf') ? blob : new Blob([blob], { type: 'application/pdf' })
  const url = URL.createObjectURL(typed)
  const win = window.open(url, '_blank')
  if (!win) {
    const a = document.createElement('a')
    a.href = url
    a.download = 'cr.pdf'
    a.rel = 'noopener'
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
    return 'download'
  }
  window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
  return 'tab'
}
