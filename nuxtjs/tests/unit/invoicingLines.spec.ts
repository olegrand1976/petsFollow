import { describe, expect, it } from 'vitest'
import {
  buildInvoicePayloadLines,
  defaultVatPercent,
  formatInvoiceMoney,
  isInvoiceAddressRequired,
  isInvoiceVatRequired,
  lineAmountsCents,
  sumInvoiceLineTotals,
} from '../../utils/invoicing-lines'

describe('defaultVatPercent', () => {
  it('applique 22 % pour l’Italie, 21 % ailleurs', () => {
    expect(defaultVatPercent('IT')).toBe(22)
    expect(defaultVatPercent('BE')).toBe(21)
    expect(defaultVatPercent('FR')).toBe(21)
    expect(defaultVatPercent('ES')).toBe(21)
  })
})

describe('isInvoiceAddressRequired / isInvoiceVatRequired', () => {
  it('exige l’adresse pour un particulier quel que soit le pays', () => {
    expect(isInvoiceAddressRequired('individual', 'FR')).toBe(true)
    expect(isInvoiceAddressRequired('individual', 'ES')).toBe(true)
  })

  it('exige l’adresse pour un pro belge uniquement', () => {
    expect(isInvoiceAddressRequired('business', 'BE')).toBe(true)
    expect(isInvoiceAddressRequired('business', 'FR')).toBe(false)
    expect(isInvoiceAddressRequired('business', 'IT')).toBe(false)
  })

  it('exige la TVA pro sur BE/FR/IT, pas ES ni particulier', () => {
    expect(isInvoiceVatRequired('business', 'BE')).toBe(true)
    expect(isInvoiceVatRequired('business', 'FR')).toBe(true)
    expect(isInvoiceVatRequired('business', 'IT')).toBe(true)
    expect(isInvoiceVatRequired('business', 'ES')).toBe(false)
    expect(isInvoiceVatRequired('individual', 'BE')).toBe(false)
  })
})

describe('lineAmountsCents', () => {
  it('calcule HT / TVA / TTC en centimes (arrondi ligne)', () => {
    expect(lineAmountsCents({
      description: 'Consult',
      quantity: '1',
      unitPriceExcl: '42',
      vatPercent: 21,
    })).toEqual({ excl: 4200, vat: 882, incl: 5082 })
  })

  it('multiplie par la quantité avant TVA', () => {
    expect(lineAmountsCents({
      description: 'Acte',
      quantity: '2',
      unitPriceExcl: '10',
      vatPercent: 6,
    })).toEqual({ excl: 2000, vat: 120, incl: 2120 })
  })

  it('accepte un taux à 0 %', () => {
    expect(lineAmountsCents({
      description: 'Exonéré',
      quantity: 1,
      unitPriceExcl: 10,
      vatPercent: 0,
    })).toEqual({ excl: 1000, vat: 0, incl: 1000 })
  })

  it('ignore qty / PU / TVA invalides ou description vide', () => {
    expect(lineAmountsCents({ description: 'X', quantity: '0', unitPriceExcl: '10', vatPercent: 21 })).toBeNull()
    expect(lineAmountsCents({ description: 'X', quantity: '1', unitPriceExcl: '', vatPercent: 21 })).toBeNull()
    expect(lineAmountsCents({ description: 'X', quantity: '1', unitPriceExcl: '-5', vatPercent: 21 })).toBeNull()
    expect(lineAmountsCents({ description: 'X', quantity: '1', unitPriceExcl: '10', vatPercent: Number.NaN })).toBeNull()
    // Prix saisi sans libellé : hors totaux et hors payload (évite un TTC trompeur).
    expect(lineAmountsCents({ description: '  ', quantity: '1', unitPriceExcl: '50', vatPercent: 21 })).toBeNull()
  })

  it('convertit les euros décimaux en centimes sans dérive', () => {
    // 12.50 € → 1250 cents (éviter 12.499999)
    expect(lineAmountsCents({
      description: 'Médicament',
      quantity: '3',
      unitPriceExcl: '12.50',
      vatPercent: 6,
    })).toEqual({ excl: 3750, vat: 225, incl: 3975 })
  })
})

describe('sumInvoiceLineTotals', () => {
  // Miroir du scénario e2e I7.1 : 42 € @21 % + 2×10 € @6 % → 62.00 HT / 72.02 TTC
  it('agrège le scénario multi-lignes de la suite e2e', () => {
    const totals = sumInvoiceLineTotals([
      { description: 'Consultation E2E', quantity: '1', unitPriceExcl: '42', vatPercent: 21 },
      { description: 'Acte complémentaire', quantity: '2', unitPriceExcl: '10', vatPercent: 6 },
    ])
    expect(totals).toEqual({ excl: 6200, vat: 1002, incl: 7202 })
    expect(formatInvoiceMoney(totals.excl)).toBe('62.00 €')
    expect(formatInvoiceMoney(totals.incl)).toBe('72.02 €')
  })

  it('ignore les lignes incomplètes sans fausser le total', () => {
    const totals = sumInvoiceLineTotals([
      { description: 'OK', quantity: '1', unitPriceExcl: '100', vatPercent: 21 },
      { description: '', quantity: '1', unitPriceExcl: '50', vatPercent: 21 },
      { description: 'skip', quantity: '0', unitPriceExcl: '50', vatPercent: 21 },
    ])
    expect(totals.excl).toBe(10000)
    expect(totals.vat).toBe(2100)
    expect(totals.incl).toBe(12100)
  })
})

describe('buildInvoicePayloadLines', () => {
  it('émet le body attendu par l’API (centimes + description trim)', () => {
    expect(buildInvoicePayloadLines([
      { description: '  Consult  ', quantity: '1', unitPriceExcl: '45.00', vatPercent: 21 },
      { description: '', quantity: '1', unitPriceExcl: '10', vatPercent: 21 },
      { description: 'Vide PU', quantity: '1', unitPriceExcl: '', vatPercent: 21 },
    ])).toEqual([
      {
        description: 'Consult',
        quantity: 1,
        unitPriceExclCents: 4500,
        vatPercent: 21,
      },
    ])
  })

  it('refuse un document sans aucune ligne valide', () => {
    expect(buildInvoicePayloadLines([
      { description: '', quantity: '1', unitPriceExcl: '10', vatPercent: 21 },
    ])).toEqual([])
  })
})
