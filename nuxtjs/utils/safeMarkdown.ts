import { marked } from 'marked'
import DOMPurify from 'isomorphic-dompurify'

marked.setOptions({ breaks: true, gfm: true })

/** Strip NULs and normalize unicode before markdown render / API payloads. */
export function normalizeReportText(raw: string): string {
  if (!raw) return ''
  // Drop C0 controls except tab/LF/CR.
  // eslint-disable-next-line no-control-regex
  const cleaned = raw.replace(/[\u0000-\u0008\u000B\u000C\u000E-\u001F]/g, '')
  try {
    return cleaned.normalize('NFKC')
  }
  catch {
    return cleaned
  }
}

/** Markdown → sanitized HTML for CR preview (never trust model output). */
export function renderSafeMarkdown(source: string): string {
  const text = normalizeReportText(source)
  if (!text.trim()) return ''
  const html = marked.parse(text, { async: false }) as string
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input', 'button'],
    FORBID_ATTR: ['onerror', 'onclick', 'onload', 'style'],
  })
}
