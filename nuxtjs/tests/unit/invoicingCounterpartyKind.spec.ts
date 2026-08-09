import { describe, it, expect } from 'vitest'
import { counterpartyKind } from '../../utils/invoicing-counterparty'

describe('counterpartyKind', () => {
  it('respecte le type explicite', () => {
    expect(counterpartyKind({ customerKind: 'business' })).toBe('business')
    expect(counterpartyKind({ customerKind: 'individual', vatNumber: 'BE1000000021' })).toBe('individual')
  })

  // Avoir sur une facture antérieure à la bascule : sans identifiant fiscal
  // conservé, la TVA sauterait du document et l'avoir partirait par email.
  it('déduit business d’une contrepartie ancienne portant un identifiant fiscal', () => {
    expect(counterpartyKind({ vatNumber: 'BE1000000021' })).toBe('business')
    expect(counterpartyKind({ companyNumber: '1000000021' })).toBe('business')
    expect(counterpartyKind({ siret: '12345678901234' })).toBe('business')
    expect(counterpartyKind({ siren: '123456789' })).toBe('business')
    expect(counterpartyKind({ codiceDestinatario: 'ABCDEFG' })).toBe('business')
    expect(counterpartyKind({ pec: 'studio@pec.it' })).toBe('business')
    expect(counterpartyKind({ taxId: 'B12345678' })).toBe('business')
  })

  it('déduit particulier sans identifiant fiscal', () => {
    expect(counterpartyKind({})).toBe('individual')
    expect(counterpartyKind({ vatNumber: '', companyNumber: '' })).toBe('individual')
  })

  it('ignore une valeur inconnue et retombe sur la déduction', () => {
    expect(counterpartyKind({ customerKind: 'particulier', vatNumber: 'BE1000000021' })).toBe('business')
    expect(counterpartyKind({ customerKind: 'particulier' })).toBe('individual')
  })
})
