/** Client-side helpers for Compendium import review (dual lists + filters + pagination). */

export type CompendiumRowLike = {
  id?: string
  status?: string
  cnk?: string
  name?: string
  manufacturer?: string
  activeSubstance?: string
  strength?: string
  atcCode?: string
  suggestedCnk?: string
  pharmaceuticalForm?: string
  packSize?: string
  errorCode?: string
  rowNumber?: number
}

export type CompendiumReviewStatusFilter = 'all' | 'pending' | 'error'

export type CompendiumReviewFilter = {
  q: string
  status: CompendiumReviewStatusFilter
  hasSuggested: boolean
}

export type CompendiumReviewTotals = {
  toReview: number
  ready: number
  excluded: number
  upserted: number
  loaded: number
}

export const COMPENDIUM_REVIEW_PAGE_SIZE = 20

export function isReviewQueueRow (row: CompendiumRowLike): boolean {
  return row.status === 'pending' || row.status === 'error'
}

export function isReadyQueueRow (row: CompendiumRowLike): boolean {
  return row.status === 'ready'
}

export function isExcludedQueueRow (row: CompendiumRowLike): boolean {
  return row.status === 'excluded'
}

export function compendiumReviewTotals (rows: CompendiumRowLike[]): CompendiumReviewTotals {
  let toReview = 0
  let ready = 0
  let excluded = 0
  let upserted = 0
  for (const r of rows) {
    if (isReviewQueueRow(r)) toReview++
    else if (isReadyQueueRow(r)) ready++
    else if (r.status === 'excluded') excluded++
    else if (r.status === 'upserted') upserted++
  }
  return { toReview, ready, excluded, upserted, loaded: rows.length }
}

function rowHaystack (row: CompendiumRowLike): string {
  return [
    row.cnk,
    row.name,
    row.manufacturer,
    row.activeSubstance,
    row.strength,
    row.atcCode,
    row.suggestedCnk,
    row.pharmaceuticalForm,
    row.packSize,
    row.errorCode,
    row.rowNumber != null ? String(row.rowNumber) : '',
  ]
    .map(v => String(v || '').toLowerCase())
    .join(' ')
}

export function filterCompendiumReviewRows (
  rows: CompendiumRowLike[],
  filter: CompendiumReviewFilter,
): CompendiumRowLike[] {
  const q = filter.q.trim().toLowerCase()
  return rows.filter((row) => {
    if (filter.status === 'pending' && row.status !== 'pending') return false
    if (filter.status === 'error' && row.status !== 'error') return false
    if (filter.hasSuggested && !String(row.suggestedCnk || '').trim()) return false
    if (q && !rowHaystack(row).includes(q)) return false
    return true
  })
}

export function paginateRows<T> (
  rows: T[],
  page: number,
  pageSize: number = COMPENDIUM_REVIEW_PAGE_SIZE,
): { items: T[]; page: number; pageCount: number; from: number; to: number; total: number } {
  const size = Math.max(1, Math.trunc(pageSize) || COMPENDIUM_REVIEW_PAGE_SIZE)
  const total = rows.length
  const pageCount = Math.max(1, Math.ceil(total / size) || 1)
  const safePage = Math.min(Math.max(1, Math.trunc(page) || 1), pageCount)
  const start = (safePage - 1) * size
  const items = rows.slice(start, start + size)
  const from = total === 0 ? 0 : start + 1
  const to = total === 0 ? 0 : start + items.length
  return { items, page: safePage, pageCount, from, to, total }
}
