import { describe, expect, it } from 'vitest'
import { prefillDraft } from '../../utils/invoicing-prefill'

/**
 * Ces lignes partent chez le client une fois la facture envoyée : un centime mal
 * converti ou une TVA imposée par défaut ne se rattrape pas côté Billit.
 */
describe('prefillDraft', () => {
  it('convertit les centimes de l’API en euros saisissables', () => {
    const draft = prefillDraft(
      { lines: [{ description: 'Consultation', quantity: 1, unitPriceExclCents: 4500, vatPercent: 21, source: 'visitType' }] },
      21,
    )
    expect(draft.lines).toEqual([
      { description: 'Consultation', quantity: '1', unitPriceExcl: '45.00', vatPercent: 21 },
    ])
    expect(draft.estimatedExclCents).toBe(4500)
  })

  it('multiplie le total par la quantité', () => {
    const draft = prefillDraft(
      {
        lines: [
          { description: 'Acte', quantity: 1, unitPriceExclCents: 4500, vatPercent: 21, source: 'visitType' },
          { description: 'Antibiotique', quantity: 3, unitPriceExclCents: 1250, vatPercent: 6, source: 'daf' },
        ],
      },
      21,
    )
    expect(draft.estimatedExclCents).toBe(4500 + 3 * 1250)
    expect(draft.dafLineCount).toBe(1)
    expect(draft.lines[1]).toMatchObject({ quantity: '3', unitPriceExcl: '12.50', vatPercent: 6 })
  })

  it('laisse le prix vide quand le médicament n’est pas au catalogue', () => {
    const draft = prefillDraft(
      { lines: [{ description: 'Hors catalogue', quantity: 2, unitPriceExclCents: 0, source: 'daf' }] },
      21,
    )
    expect(draft.lines[0].unitPriceExcl).toBe('')
    // Un montant estimé partiel afficherait un total faux au véto.
    expect(draft.estimatedExclCents).toBe(0)
  })

  it('applique le taux du pays quand l’API ne connaît pas la TVA', () => {
    const draft = prefillDraft(
      { lines: [{ description: 'Hors catalogue', quantity: 1, source: 'daf' }] },
      22,
    )
    expect(draft.lines[0].vatPercent).toBe(22)
  })

  it('garde un taux à 0 % renvoyé explicitement par l’API', () => {
    const draft = prefillDraft(
      { lines: [{ description: 'Exonéré', quantity: 1, unitPriceExclCents: 1000, vatPercent: 0, source: 'daf' }] },
      21,
    )
    expect(draft.lines[0].vatPercent).toBe(0)
  })

  it('retombe sur une quantité de 1 si l’API en renvoie une aberrante', () => {
    const draft = prefillDraft(
      { lines: [{ description: 'Acte', quantity: 0, unitPriceExclCents: 4500, source: 'visitType' }] },
      21,
    )
    expect(draft.lines[0].quantity).toBe('1')
    expect(draft.estimatedExclCents).toBe(4500)
  })

  it('ne propose rien sur une réponse vide ou absente', () => {
    for (const payload of [null, undefined, {}, { lines: [] }]) {
      const draft = prefillDraft(payload as any, 21)
      expect(draft.lines).toHaveLength(0)
      expect(draft.estimatedExclCents).toBe(0)
      expect(draft.dafLineCount).toBe(0)
    }
  })
})
