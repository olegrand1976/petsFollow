import TurndownService from 'turndown'
import { normalizeReportText, renderSafeMarkdown } from './safeMarkdown'

const turndown = new TurndownService({
  headingStyle: 'atx',
  bulletListMarker: '-',
  codeBlockStyle: 'fenced',
  emDelimiter: '*',
  strongDelimiter: '**',
})

/** Markdown → HTML for TipTap (sanitized). */
export function markdownToEditorHtml(md: string): string {
  const html = renderSafeMarkdown(md)
  return html.trim() ? html : '<p></p>'
}

/** TipTap HTML → Markdown for API persistence. */
export function editorHtmlToMarkdown(html: string): string {
  if (!html || !html.replace(/<[^>]+>/g, '').trim()) return ''
  return normalizeReportText(turndown.turndown(html))
}

/**
 * Round-trip normalize so TipTap↔turndown whitespace/list quirks do not
 * flip dirty / re-setContent in a loop.
 */
export function canonicalizeReportMarkdown(md: string): string {
  const raw = normalizeReportText(md).replace(/\r\n/g, '\n').trim()
  if (!raw) return ''
  return editorHtmlToMarkdown(markdownToEditorHtml(raw)).replace(/\r\n/g, '\n').trim()
}
