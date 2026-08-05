/** Open a consignes PDF via BFF blob (errors instead of a blank hanging tab). */
export async function openPrescriptionPdfBlob(
  prescriptionId: string,
): Promise<'tab' | 'download'> {
  const id = String(prescriptionId || '').trim()
  if (!id) {
    const err: any = new Error('missing_id')
    err.data = { code: 'missing_id' }
    throw err
  }

  const blob = await $fetch<Blob>(`/api/vet/prescriptions/${id}/pdf`, {
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
    // Stay on the consignes page: force a download instead of same-tab navigation.
    const a = document.createElement('a')
    a.href = url
    a.download = 'consignes.pdf'
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
