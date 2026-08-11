import { describe, expect, it } from 'vitest'
import {
  formatPharmacyCents,
  isPharmacyWasteReason,
  pharmacyMovementDetailKey,
  pharmacyMovementReasonKey,
  pharmacyPriceIsPersisted,
  pharmacyWithdrawalDays,
} from '~/utils/pharmacy-stock'

describe('pharmacy-stock helpers', () => {
  it('maps movement reasons to i18n keys', () => {
    expect(pharmacyMovementReasonKey('receipt')).toBe('pharmacy.stock.reasonReceipt')
    expect(pharmacyMovementReasonKey('daf')).toBe('pharmacy.stock.reasonDaf')
    expect(pharmacyMovementReasonKey('waste')).toBe('pharmacy.stock.reasonWaste')
    expect(pharmacyMovementReasonKey('unknown')).toBe('pharmacy.stock.reasonUnknown')
  })

  it('maps reason details when known', () => {
    expect(pharmacyMovementDetailKey('auto_expired')).toBe('pharmacy.stock.detailAutoExpired')
    expect(pharmacyMovementDetailKey('manual')).toBe('pharmacy.stock.detailManual')
    expect(pharmacyMovementDetailKey('weird_code')).toBeNull()
  })

  it('validates waste reasons', () => {
    expect(isPharmacyWasteReason('expired')).toBe(true)
    expect(isPharmacyWasteReason('supplier_return')).toBe(true)
    expect(isPharmacyWasteReason('destruction')).toBe(true)
    expect(isPharmacyWasteReason('other')).toBe(false)
  })

  it('detects persisted prices', () => {
    expect(pharmacyPriceIsPersisted({ updatedAt: '2026-01-01T00:00:00Z' })).toBe(true)
    expect(pharmacyPriceIsPersisted({ purchasePriceCents: 100 })).toBe(true)
    expect(pharmacyPriceIsPersisted({ sellPriceCents: 50 })).toBe(true)
    expect(pharmacyPriceIsPersisted({ purchasePriceCents: 0, sellPriceCents: 0 })).toBe(false)
  })

  it('formats cents and normalizes withdrawal days', () => {
    expect(formatPharmacyCents(1234)).toBe('12.34 €')
    expect(formatPharmacyCents(0)).toBe('0.00 €')
    expect(pharmacyWithdrawalDays(7)).toBe(7)
    expect(pharmacyWithdrawalDays(7.9)).toBe(7)
    expect(pharmacyWithdrawalDays(null)).toBeNull()
    expect(pharmacyWithdrawalDays(undefined)).toBeNull()
    expect(pharmacyWithdrawalDays(-3)).toBe(0)
  })
})
