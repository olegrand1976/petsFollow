import { describe, expect, it } from 'vitest'
import { mapVisitReportFields, persistedHistoryBody } from '../../utils/visitReport'
import { normalizeReportText, renderSafeMarkdown } from '../../utils/safeMarkdown'
import {
  canonicalizeReportMarkdown,
  editorHtmlToMarkdown,
  markdownToEditorHtml,
} from '../../utils/reportRichText'

/** Shared escape / hostile corpus (aligned with Go normalize tests). */
export const REPORT_ESCAPE_CORPUS
  = 'Notes "quotes" \'apos\' «guillemets» back\\slash &amp; <script>alert(1)</script>\n'
    + '**bold**\n- item\n'
    + 'emoji 🐶 accents été\u0000\u0001'

describe('mapVisitReportFields', () => {
  it('returns empty fields for null payload', () => {
    expect(mapVisitReportFields(null)).toEqual({
      bodyText: '',
      transcriptText: '',
      improvedText: '',
      status: '',
      isReference: false,
    })
  })

  it('maps transcript, improved and body independently', () => {
    expect(
      mapVisitReportFields({
        bodyText: 'edited body',
        transcriptText: 'raw transcript',
        improvedText: 'ai version',
        status: 'draft',
        isReference: true,
      }),
    ).toEqual({
      bodyText: 'edited body',
      transcriptText: 'raw transcript',
      improvedText: 'ai version',
      status: 'draft',
      isReference: true,
    })
  })

  it('falls back body to transcript when body is empty', () => {
    expect(
      mapVisitReportFields({
        bodyText: '',
        transcriptText: 'from audio',
        improvedText: '',
        status: 'draft',
      }),
    ).toEqual({
      bodyText: 'from audio',
      transcriptText: 'from audio',
      improvedText: '',
      status: 'draft',
      isReference: false,
    })
  })
})

describe('persistedHistoryBody', () => {
  it('hides body identical to transcript or IA', () => {
    expect(persistedHistoryBody('same', 'same', 'ai')).toBe('')
    expect(persistedHistoryBody('ai', 'raw', 'ai')).toBe('')
  })

  it('returns distinct persisted body', () => {
    expect(persistedHistoryBody('edited', 'raw', 'ai')).toBe('edited')
  })

  it('ignores empty persisted body', () => {
    expect(persistedHistoryBody('', 'raw', 'ai')).toBe('')
    expect(persistedHistoryBody('', 'raw', 'ai')).toBe('')
  })
})

describe('reportRichText', () => {
  it('round-trips bold section headings without raw asterisks in HTML', () => {
    const md = '**Anamnèse / motif :**\n\n- item détail'
    const html = markdownToEditorHtml(md)
    expect(html).toContain('<strong>')
    expect(html).not.toContain('**Anamnèse')
    const back = editorHtmlToMarkdown(html)
    expect(back).toContain('Anamnèse')
    expect(back).toMatch(/\*\*|__/)
  })

  it('canonicalize stabilizes TipTap round-trips for dirty checks', () => {
    const md = '**Anamnèse / motif :**\n\n- item'
    const once = canonicalizeReportMarkdown(md)
    const twice = canonicalizeReportMarkdown(once)
    expect(twice).toBe(once)
  })
})

describe('safeMarkdown', () => {
  it('strips NUL and control chars', () => {
    expect(normalizeReportText('a\u0000b\u0001c')).toBe('abc')
  })

  it('preserves quotes backslash entities accents emoji', () => {
    const cleaned = normalizeReportText(REPORT_ESCAPE_CORPUS)
    expect(cleaned).not.toMatch(/\u0000|\u0001/)
    expect(cleaned).toContain('"quotes"')
    expect(cleaned).toContain('back\\slash')
    expect(cleaned).toContain('&amp;')
    expect(cleaned).toContain('été')
    expect(cleaned).toContain('🐶')
  })

  it('renders bold markdown safely', () => {
    const html = renderSafeMarkdown('**Anamnèse / motif :**\n\n- item')
    expect(html).toContain('<strong>')
    expect(html).toContain('Anamnèse')
    expect(html).toContain('<li>')
  })

  it('strips script tags from malicious markdown/html', () => {
    const html = renderSafeMarkdown('ok <script>alert(1)</script> **x**')
    expect(html.toLowerCase()).not.toContain('<script')
    expect(html).toContain('<strong>')
  })

  it('strips onerror-style payloads from corpus preview', () => {
    const html = renderSafeMarkdown(REPORT_ESCAPE_CORPUS)
    expect(html.toLowerCase()).not.toContain('<script')
    expect(html.toLowerCase()).not.toContain('onerror')
    expect(html).toContain('<strong>')
  })
})
