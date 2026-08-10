import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  copyVisitReportMarkdown,
  downloadVisitReportMarkdown,
  normalizeReportCitations,
  openVisitReportPdfBlob,
  stripReportCitations,
} from '~/utils/visit-report-export'

describe('visit-report-export', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('normalizeReportCitations accepte chaînes et objets, écarte le reste', () => {
    expect(normalizeReportCitations([
      '  Guide BSAVA  ',
      { title: 'Merck Vet Manual' },
      { title: '   ' },
      { url: 'https://exemple.test' },
      '',
      42,
      null,
    ])).toEqual(['Guide BSAVA', 'Merck Vet Manual'])
  })

  it('normalizeReportCitations rend un tableau vide hors tableau', () => {
    expect(normalizeReportCitations(undefined)).toEqual([])
    expect(normalizeReportCitations('Guide BSAVA')).toEqual([])
  })

  it('stripReportCitations removes RAG section', () => {
    const md = '## Motif\n\nToux\n\n## Références RAG\n\n- Guide BSAVA\n'
    expect(stripReportCitations(md)).toBe('## Motif\n\nToux')
  })

  it('copyVisitReportMarkdown strips when opted in', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(globalThis, 'navigator', {
      value: { clipboard: { writeText } },
      configurable: true,
    })
    await copyVisitReportMarkdown('Body\n\n## Références RAG\n\n- x\n', { stripCitations: true })
    expect(writeText).toHaveBeenCalledWith('Body')
  })

  it('copyVisitReportMarkdown writes clipboard', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(globalThis, 'navigator', {
      value: { clipboard: { writeText } },
      configurable: true,
    })
    await copyVisitReportMarkdown('## Motif\n\nToux')
    expect(writeText).toHaveBeenCalledWith('## Motif\n\nToux')
  })

  it('copyVisitReportMarkdown rejects empty', async () => {
    await expect(copyVisitReportMarkdown('   ')).rejects.toMatchObject({
      data: { error: { code: 'report_empty', msgKey: 'report_empty' } },
    })
  })

  it('copyVisitReportMarkdown signale un presse-papier indisponible', async () => {
    Object.defineProperty(globalThis, 'navigator', {
      value: {},
      configurable: true,
    })
    await expect(copyVisitReportMarkdown('# CR')).rejects.toMatchObject({
      data: { error: { msgKey: 'clipboard_unavailable' } },
    })
  })

  it('downloadVisitReportMarkdown rejects empty', () => {
    expect(() => downloadVisitReportMarkdown('')).toThrowError(/report_empty/)
  })

  it('downloadVisitReportMarkdown creates blob link', () => {
    const click = vi.fn()
    const remove = vi.fn()
    const createElement = vi.fn(() => {
      const el: Record<string, unknown> = {
        href: '',
        download: '',
        rel: '',
        click,
        remove,
      }
      return el
    })
    const appendChild = vi.fn()
    const createObjectURL = vi.fn(() => 'blob:md')
    const revokeObjectURL = vi.fn()
    vi.stubGlobal('document', {
      createElement,
      body: { appendChild },
    })
    vi.stubGlobal('URL', { createObjectURL, revokeObjectURL })
    vi.stubGlobal('window', { setTimeout: (fn: () => void) => fn() })

    downloadVisitReportMarkdown('# CR', 'Rex')
    expect(createObjectURL).toHaveBeenCalled()
    expect(createElement).toHaveBeenCalledWith('a')
    expect(appendChild).toHaveBeenCalled()
    expect(click).toHaveBeenCalled()
    expect(remove).toHaveBeenCalled()
    expect(createElement.mock.results[0]?.value.download).toBe('Rex.md')
  })
})

describe('openVisitReportPdfBlob', () => {
  /** Décor minimal du navigateur ; `open` décide onglet ou téléchargement. */
  function stubBrowser(open: () => unknown) {
    const click = vi.fn()
    const remove = vi.fn()
    const anchor: Record<string, unknown> = { href: '', download: '', rel: '', click, remove }
    vi.stubGlobal('document', {
      createElement: vi.fn(() => anchor),
      body: { appendChild: vi.fn() },
    })
    vi.stubGlobal('URL', { createObjectURL: vi.fn(() => 'blob:pdf'), revokeObjectURL: vi.fn() })
    vi.stubGlobal('window', { open: vi.fn(open), setTimeout: vi.fn() })
    return { anchor, click }
  }

  beforeEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('refuse un identifiant vide sans appeler la BFF', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('$fetch', fetchMock)
    await expect(openVisitReportPdfBlob('  ')).rejects.toMatchObject({
      data: { error: { msgKey: 'missing_id' } },
    })
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('ouvre le PDF dans un onglet', async () => {
    vi.stubGlobal('$fetch', vi.fn(async () => new Blob(['%PDF-1.4'], { type: 'application/pdf' })))
    stubBrowser(() => ({}))
    await expect(openVisitReportPdfBlob('v-1')).resolves.toBe('tab')
  })

  it('retombe sur le téléchargement quand le popup est bloqué', async () => {
    vi.stubGlobal('$fetch', vi.fn(async () => new Blob(['%PDF-1.4'], { type: 'application/pdf' })))
    const { anchor, click } = stubBrowser(() => null)
    await expect(openVisitReportPdfBlob('v-1')).resolves.toBe('download')
    expect(click).toHaveBeenCalled()
    expect(anchor.download).toBe('cr.pdf')
  })

  // La BFF peut répondre 200 avec un corps JSON (erreur amont non typée) : sans
  // cette détection, le véto ouvrirait un onglet sur un PDF illisible.
  it('transforme un corps JSON renvoyé en 200 en erreur typée', async () => {
    const payload = JSON.stringify({ error: { code: 'report_not_found' } })
    vi.stubGlobal('$fetch', vi.fn(async () => new Blob([payload], { type: 'application/json' })))
    const { click } = stubBrowser(() => ({}))
    await expect(openVisitReportPdfBlob('v-1')).rejects.toMatchObject({
      data: { error: { msgKey: 'report_not_found' } },
    })
    expect(click).not.toHaveBeenCalled()
  })

  it('retype un blob sans content-type en PDF', async () => {
    vi.stubGlobal('$fetch', vi.fn(async () => new Blob(['%PDF-1.4'])))
    stubBrowser(() => ({}))
    await expect(openVisitReportPdfBlob('v-1')).resolves.toBe('tab')
  })
})
