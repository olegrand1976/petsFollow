import { describe, expect, it } from 'vitest'
import { isRedundantDelivery, parseDeliveryStatus } from '../../utils/invoicing-delivery-status'
import { LIVE_LOCALES, readLocale } from './support/locales'

function deliveryCatalogue(locale: string): Record<string, string> {
  return readLocale(locale).invoicing.deliveryStatus
}

describe('parseDeliveryStatus', () => {
  it('ne renvoie rien tant que le document n\'est pas parti', () => {
    expect(parseDeliveryStatus('')).toBeNull()
    expect(parseDeliveryStatus('   ')).toBeNull()
    expect(parseDeliveryStatus(undefined)).toBeNull()
    expect(parseDeliveryStatus(null)).toBeNull()
  })

  it('traduit un statut réseau', () => {
    expect(parseDeliveryStatus('delivered')).toEqual({
      raw: 'delivered', byEmail: false, base: 'delivered', key: 'invoicing.deliveryStatus.delivered',
    })
  })

  // `base` sert à taire la mention quand elle répéterait le statut du document
  // (« Livrée · Remise confirmée ») : elle doit ignorer le canal.
  it('isole le canal email du statut de base', () => {
    expect(parseDeliveryStatus('email_delivered')).toEqual({
      raw: 'email_delivered', byEmail: true, base: 'delivered', key: 'invoicing.deliveryStatus.delivered',
    })
  })

  // Un statut ajouté côté Go doit dégrader vers la valeur technique : afficher
  // une clé i18n brute (« invoicing.deliveryStatus.foo ») serait pire que rien.
  it('retombe sur la valeur brute pour un statut inconnu', () => {
    expect(parseDeliveryStatus('quantum_flux')).toEqual({
      raw: 'quantum_flux', byEmail: false, base: 'quantum_flux', key: null,
    })
    expect(parseDeliveryStatus('email_quantum_flux')).toEqual({
      raw: 'email_quantum_flux', byEmail: true, base: 'quantum_flux', key: null,
    })
  })
})

describe('isRedundantDelivery', () => {
  const show = (raw: string, docStatus: string) => {
    const status = parseDeliveryStatus(raw)!
    return !isRedundantDelivery(status, docStatus)
  }

  it('tait la mention qui répète le statut du document', () => {
    expect(show('delivered', 'delivered')).toBe(false)
    expect(show('rejected', 'rejected')).toBe(false)
    // Proforma acceptée : le Go pose les deux à `accepted`, d'où l'absence de
    // libellé traduit pour ce statut.
    expect(show('accepted', 'accepted')).toBe(false)
  })

  it('garde la mention quand elle apporte le canal ou un état distinct', () => {
    expect(show('email_delivered', 'delivered')).toBe(true)
    expect(show('stale_timeout', 'sending')).toBe(true)
    expect(show('awaiting_client', 'issued')).toBe(true)
    expect(show('creating_order', 'draft')).toBe(true)
    expect(show('quantum_flux', 'sending')).toBe(true)
  })
})

/**
 * Statuts écrits par le Go (`service.go`, `store/invoicing.go`,
 * `billit/webhook.go`) **et** susceptibles d'atteindre la liste d'un cabinet.
 * Si l'un disparaît du catalogue, le véto relit du technique.
 *
 * Deux exclusions assumées : `saas_draft`, que `ListDocuments` écarte via
 * `source = 'practice'`, et `accepted`, écrit en même temps que le statut de
 * document homonyme donc toujours tu par `isRedundantDelivery`. Les traduire
 * reviendrait à maintenir douze libellés que personne ne peut lire.
 */
const GO_STATUSES = [
  'creating_order', 'sending', 'delivered', 'rejected',
  'cancelled', 'unknown', 'send_failed', 'stale_timeout',
  'awaiting_client',
]

describe('catalogue i18n des statuts d\'acheminement', () => {
  it.each(LIVE_LOCALES)('%s couvre tous les statuts produits par l\'API', (locale) => {
    const cat = deliveryCatalogue(locale)
    for (const status of GO_STATUSES) {
      expect(cat[status], `${locale}: ${status} manquant`).toBeTruthy()
    }
    expect(cat.byEmail).toContain('{status}')
  })

  it('le helper reconnaît exactement les statuts du catalogue', () => {
    const known = Object.keys(deliveryCatalogue('fr')).filter((k) => k !== 'byEmail')
    for (const status of known) {
      expect(parseDeliveryStatus(status)?.key).toBe(`invoicing.deliveryStatus.${status}`)
    }
    expect(known.sort()).toEqual([...GO_STATUSES].sort())
  })
})
