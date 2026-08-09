import { describe, expect, it } from 'vitest'
import {
  isValidVisitTypeVat,
  visitTypePriceCents,
  visitTypePriceInput,
} from '~/utils/visit-type-pricing'

describe('visitTypePriceCents', () => {
  it('convertit les euros saisis en centimes', () => {
    expect(visitTypePriceCents('45')).toBe(4500)
    expect(visitTypePriceCents('45.50')).toBe(4550)
    expect(visitTypePriceCents(' 45.50 ')).toBe(4550)
  })

  it('accepte la virgule décimale (clavier FR/NL)', () => {
    expect(visitTypePriceCents('45,50')).toBe(4550)
  })

  it('arrondit au centime au lieu de tronquer', () => {
    expect(visitTypePriceCents('12.345')).toBe(1235)
    expect(visitTypePriceCents('0.005')).toBe(1)
  })

  it('traite le champ vide comme « non tarifé » et non comme une erreur', () => {
    expect(visitTypePriceCents('')).toBe(0)
    expect(visitTypePriceCents('   ')).toBe(0)
    expect(visitTypePriceCents(null)).toBe(0)
    expect(visitTypePriceCents(undefined)).toBe(0)
  })

  it('refuse un montant négatif ou non numérique', () => {
    expect(visitTypePriceCents('-1')).toBeNull()
    expect(visitTypePriceCents('gratuit')).toBeNull()
  })
})

describe('visitTypePriceInput', () => {
  it('affiche un tarif à deux décimales', () => {
    expect(visitTypePriceInput(4500)).toBe('45.00')
    expect(visitTypePriceInput(4550)).toBe('45.50')
  })

  it('laisse le champ vide quand rien n’est tarifé', () => {
    expect(visitTypePriceInput(0)).toBe('')
    expect(visitTypePriceInput(undefined)).toBe('')
    expect(visitTypePriceInput('abc')).toBe('')
  })

  it('fait l’aller-retour saisie → API → saisie sans dérive', () => {
    const cents = visitTypePriceCents('45,50')
    expect(visitTypePriceInput(cents)).toBe('45.50')
  })
})

describe('isValidVisitTypeVat', () => {
  it('accepte les taux du catalogue de facturation', () => {
    for (const vat of [0, 6, 12, 21, 22]) {
      expect(isValidVisitTypeVat(vat)).toBe(true)
    }
  })

  it('refuse hors bornes — l’API rejetterait la ligne de facture', () => {
    expect(isValidVisitTypeVat(-1)).toBe(false)
    expect(isValidVisitTypeVat(150)).toBe(false)
    expect(isValidVisitTypeVat('vingt et un')).toBe(false)
  })
})
