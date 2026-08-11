import { describe, expect, it } from 'vitest'
import {
  COMPENDIUM_EXTRACT_STALE_MS,
  COMPENDIUM_REVIEW_PAGE_SIZE,
  compendiumReviewTotals,
  filterCompendiumReviewRows,
  isCompendiumExtractClaimLive,
  isExcludedQueueRow,
  isReadyQueueRow,
  isReviewQueueRow,
  paginateRows,
} from '~/utils/compendium-review'

describe('compendium-review helpers', () => {
  const rows = [
    { id: '1', status: 'pending', name: 'Alpha', cnk: '111', suggestedCnk: '111' },
    { id: '2', status: 'error', name: 'Beta', cnk: '', errorCode: 'missing_cnk' },
    { id: '3', status: 'ready', name: 'Gamma', cnk: '333' },
    { id: '4', status: 'excluded', name: 'Delta', cnk: '444', atcCode: 'QJ01CA04', strength: '500 mg' },
    { id: '5', status: 'upserted', name: 'Epsilon', cnk: '555' },
  ]

  it('classifie les files revue / prêtes / exclues', () => {
    expect(rows.filter(isReviewQueueRow)).toHaveLength(2)
    expect(rows.filter(isReadyQueueRow)).toHaveLength(1)
    expect(rows.filter(isExcludedQueueRow)).toHaveLength(1)
  })

  it('calcule les totaux toolbar', () => {
    expect(compendiumReviewTotals(rows)).toEqual({
      toReview: 2,
      ready: 1,
      excluded: 1,
      upserted: 1,
      loaded: 5,
    })
  })

  it('filtre recherche + statut + CNK suggéré', () => {
    expect(filterCompendiumReviewRows(rows, { q: 'beta', status: 'all', hasSuggested: false }).map(r => r.id)).toEqual(['2'])
    expect(filterCompendiumReviewRows(rows, { q: '', status: 'error', hasSuggested: false }).map(r => r.id)).toEqual(['2'])
    expect(filterCompendiumReviewRows(rows, { q: '', status: 'all', hasSuggested: true }).map(r => r.id)).toEqual(['1'])
    expect(filterCompendiumReviewRows(rows, { q: 'qj01', status: 'all', hasSuggested: false }).map(r => r.id)).toEqual(['4'])
    expect(filterCompendiumReviewRows(rows, { q: '500 mg', status: 'all', hasSuggested: false }).map(r => r.id)).toEqual(['4'])
  })

  it('paginate avec plage X–Y / N', () => {
    const many = Array.from({ length: 25 }, (_, i) => ({ id: String(i + 1) }))
    const p1 = paginateRows(many, 1, COMPENDIUM_REVIEW_PAGE_SIZE)
    expect(p1.items).toHaveLength(20)
    expect(p1).toMatchObject({ page: 1, pageCount: 2, from: 1, to: 20, total: 25 })
    const p2 = paginateRows(many, 2, COMPENDIUM_REVIEW_PAGE_SIZE)
    expect(p2.items).toHaveLength(5)
    expect(p2).toMatchObject({ page: 2, from: 21, to: 25, total: 25 })
    expect(paginateRows([], 3).page).toBe(1)
  })

  it('détecte extract live vs stale (15 min)', () => {
    const now = Date.UTC(2026, 7, 11, 12, 0, 0)
    expect(isCompendiumExtractClaimLive('failed', new Date(now).toISOString(), now)).toBe(false)
    expect(isCompendiumExtractClaimLive('extracting', undefined, now)).toBe(true)
    expect(isCompendiumExtractClaimLive('extracting', new Date(now - 60_000).toISOString(), now)).toBe(true)
    expect(isCompendiumExtractClaimLive(
      'extracting',
      new Date(now - COMPENDIUM_EXTRACT_STALE_MS - 1).toISOString(),
      now,
    )).toBe(false)
  })
})
