/** Movement reason codes from pharmacy.stock_movements.reason */
export type PharmacyMovementReason =
  | 'receipt'
  | 'daf'
  | 'adjust'
  | 'waste'
  | 'daf_cancel'
  | 'quarantine'
  | 'unquarantine'

export const PHARMACY_WASTE_REASONS = ['expired', 'supplier_return', 'destruction'] as const
export type PharmacyWasteReason = (typeof PHARMACY_WASTE_REASONS)[number]

export function pharmacyMovementReasonKey(reason: string): string {
  const map: Record<string, string> = {
    receipt: 'pharmacy.stock.reasonReceipt',
    daf: 'pharmacy.stock.reasonDaf',
    adjust: 'pharmacy.stock.reasonAdjust',
    waste: 'pharmacy.stock.reasonWaste',
    daf_cancel: 'pharmacy.stock.reasonDafCancel',
    quarantine: 'pharmacy.stock.reasonQuarantine',
    unquarantine: 'pharmacy.stock.reasonUnquarantine',
  }
  return map[reason] || 'pharmacy.stock.reasonUnknown'
}

/** Maps stock_movements.reason_detail codes to i18n keys when known. */
export function pharmacyMovementDetailKey(detail: string): string | null {
  const map: Record<string, string> = {
    auto_expired: 'pharmacy.stock.detailAutoExpired',
    manual: 'pharmacy.stock.detailManual',
    fefo_allocate: 'pharmacy.stock.detailFefoAllocate',
    expired: 'pharmacy.stock.wasteReasonExpired',
    supplier_return: 'pharmacy.stock.wasteReasonSupplierReturn',
    destruction: 'pharmacy.stock.wasteReasonDestruction',
  }
  return map[detail] || null
}

export function isPharmacyWasteReason(value: string): value is PharmacyWasteReason {
  return (PHARMACY_WASTE_REASONS as readonly string[]).includes(value)
}

/** True when a GET price payload represents a persisted catalogue row. */
export function pharmacyPriceIsPersisted(price: {
  updatedAt?: string
  purchasePriceCents?: number
  sellPriceCents?: number
}): boolean {
  if (price.updatedAt && String(price.updatedAt).trim()) return true
  return (price.purchasePriceCents ?? 0) > 0 || (price.sellPriceCents ?? 0) > 0
}

/** Format catalogue cents as euro string (UI Pro pharmacy). */
export function formatPharmacyCents(cents: number): string {
  return `${(Number(cents) / 100).toFixed(2)} €`
}

/** Normalize nullable withdrawal day fields from API for form inputs. */
export function pharmacyWithdrawalDays(value: number | null | undefined): number | null {
  if (value === null || value === undefined || Number.isNaN(Number(value))) return null
  return Math.max(0, Math.trunc(Number(value)))
}
