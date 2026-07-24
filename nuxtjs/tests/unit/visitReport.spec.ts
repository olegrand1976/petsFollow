import { describe, expect, it } from 'vitest'
import { mapVisitReportFields, persistedHistoryBody } from '../../utils/visitReport'

describe('mapVisitReportFields', () => {
  it('returns empty fields for null payload', () => {
    expect(mapVisitReportFields(null)).toEqual({
      bodyText: '',
      transcriptText: '',
      improvedText: '',
      status: '',
    })
  })

  it('maps transcript, improved and body independently', () => {
    expect(
      mapVisitReportFields({
        bodyText: 'edited body',
        transcriptText: 'raw transcript',
        improvedText: 'ai version',
        status: 'draft',
      }),
    ).toEqual({
      bodyText: 'edited body',
      transcriptText: 'raw transcript',
      improvedText: 'ai version',
      status: 'draft',
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
  })
})
