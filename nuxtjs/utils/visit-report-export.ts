/** Client-side export helpers for visit CR (markdown + PDF). */

/**
 * Erreur locale à la forme que lit `mapError` (`data.error.msgKey` → `errors.*`).
 * Sans cette enveloppe, l'échec sort en « Une erreur est survenue ». `message`
 * porte le texte déjà traduit par l'API quand le code n'a pas de clé côté Pro.
 */
function exportError(code: string, message?: string): Error {
  const err: any = new Error(message || code)
  err.data = { error: { code, msgKey: code, ...(message ? { message } : {}) } }
  return err
}

/** Strip RAG reference sections from CR markdown (patient-safe export). */
export function stripReportCitations(md: string): string {
  let s = String(md || '').replace(/\r\n/g, '\n').replace(/\r/g, '\n')
  // ATX / plain headings for RAG refs (FR/EN/NL/ES/IT/ET + generic).
  s = s.replace(
    /\n{0,2}#{1,6}\s*(Références\s*RAG|References\s*RAG|RAG\s*(sources|references|bronnen|fuentes|fonti|allikad)|Sources\s*RAG)\s*\n[\s\S]*$/i,
    '',
  )
  // Trailing bullet list after a lone "Références" line without ATX.
  s = s.replace(
    /\n{1,2}(Références|References|Bronnen|Fuentes|Fonti|Allikad)\s*:?\s*\n(?:\s*[-*•].*\n?)+\s*$/i,
    '',
  )
  return s.replace(/\n{3,}/g, '\n\n').trim()
}

/**
 * Sources du dernier « Améliorer IA avancé ». Le Go renvoie des chaînes, CrewAI
 * peut remonter des objets `{ title }` : tout le reste est écarté plutôt
 * qu'affiché sous forme d'objet brut au véto.
 */
export function normalizeReportCitations(raw: unknown): string[] {
  if (!Array.isArray(raw)) return []
  return raw
    .map((item) => {
      if (typeof item === 'string') return item.trim()
      if (item && typeof item === 'object' && 'title' in item) {
        return String((item as { title?: unknown }).title || '').trim()
      }
      return ''
    })
    .filter(Boolean)
}

export async function copyVisitReportMarkdown(text: string, opts?: { stripCitations?: boolean }): Promise<void> {
  let body = String(text || '')
  if (opts?.stripCitations) body = stripReportCitations(body)
  if (!body.trim()) {
    throw exportError('report_empty')
  }
  if (typeof navigator === 'undefined' || !navigator.clipboard?.writeText) {
    throw exportError('clipboard_unavailable')
  }
  await navigator.clipboard.writeText(body)
}

export function downloadVisitReportMarkdown(
  text: string,
  filenameBase = 'cr',
  opts?: { stripCitations?: boolean },
): void {
  let body = String(text || '')
  if (opts?.stripCitations) body = stripReportCitations(body)
  if (!body.trim()) {
    throw exportError('report_empty')
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
  opts?: { stripCitations?: boolean },
): Promise<'tab' | 'download'> {
  const id = String(visitId || '').trim()
  if (!id) {
    throw exportError('missing_id')
  }

  const qs = opts?.stripCitations ? '?stripCitations=1' : ''
  const blob = await ($fetch as any)(`/api/visits/${id}/report-pdf${qs}`, {
    responseType: 'blob',
  }) as Blob

  const type = (blob.type || '').toLowerCase()
  if (type.includes('json') || type.includes('text/html') || type.includes('text/plain')) {
    // Code et message sont distincts : ranger une phrase dans `msgKey` la ferait
    // chercher comme clé i18n, échouer, et sortir en « Une erreur est survenue ».
    let code = 'pdf_failed'
    let message = ''
    try {
      const text = await blob.text()
      const parsed = JSON.parse(text) as {
        error?: { code?: string, message?: string }
        statusMessage?: string
        data?: { code?: string }
      }
      code = parsed?.error?.code || parsed?.data?.code || code
      message = parsed?.error?.message || parsed?.statusMessage || ''
    }
    catch { /* keep default */ }
    throw exportError(code, message)
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
