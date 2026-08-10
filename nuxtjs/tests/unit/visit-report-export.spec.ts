import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  copyVisitReportMarkdown,
  downloadVisitReportMarkdown,
} from '~/utils/visit-report-export'

describe('visit-report-export', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
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
      data: { code: 'report_empty' },
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
