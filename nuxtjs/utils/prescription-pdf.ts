/** Open a consignes / DAF-style PDF via BFF blob (shows errors instead of a blank tab). */
export async function openPrescriptionPdfBlob(prescriptionId: string): Promise<void> {
  const id = String(prescriptionId || '').trim()
  if (!id) throw new Error('missing_id')

  const blob = await $fetch<Blob>(`/api/vet/prescriptions/${id}/pdf`, {
    responseType: 'blob',
  })

  const type = (blob.type || '').toLowerCase()
  if (type.includes('json') || type.includes('text/html') || type.includes('text/plain')) {
    let msg = 'pdf_failed'
    try {
      const text = await blob.text()
      const parsed = JSON.parse(text) as { error?: { code?: string, message?: string }, statusMessage?: string }
      msg = parsed?.error?.code || parsed?.error?.message || parsed?.statusMessage || msg
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
    // Popup blocked: fall back to same-tab navigation.
    window.location.assign(url)
    return
  }
  // Revoke after the viewer had time to load the blob.
  window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
}
