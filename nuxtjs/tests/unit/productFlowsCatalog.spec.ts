import { describe, expect, it } from 'vitest'
import { parseFlowProfileId, productFlowsCatalog } from '../../data/product-flows/catalog'
import { productFlowsNavItem } from '../../utils/productFlowsPath'
import fr from '../../locales/fr.json'
import en from '../../locales/en.json'
import nl from '../../locales/nl.json'
import es from '../../locales/es.json'
import et from '../../locales/et.json'
import itLocale from '../../locales/it.json'

const locales = { fr, en, nl, es, et, it: itLocale } as const

describe('productFlows catalog', () => {
  it('expose des profils, étapes et liens croisés uniques', () => {
    expect(productFlowsCatalog.profiles.length).toBeGreaterThanOrEqual(7)
    expect(productFlowsCatalog.links.length).toBeGreaterThanOrEqual(10)

    const profileIds = productFlowsCatalog.profiles.map((p) => p.id)
    expect(new Set(profileIds).size).toBe(profileIds.length)

    const linkIds = productFlowsCatalog.links.map((l) => l.id)
    expect(new Set(linkIds).size).toBe(linkIds.length)

    for (const p of productFlowsCatalog.profiles) {
      expect(p.stepIds.length).toBeGreaterThan(0)
      expect(p.featureIds.length).toBeGreaterThan(0)
    }
    for (const link of productFlowsCatalog.links) {
      expect(profileIds).toContain(link.fromProfile)
      expect(profileIds).toContain(link.toProfile)
    }
  })

  it('référence des IDs présents dans les 6 locales', () => {
    for (const [loc, data] of Object.entries(locales)) {
      const pf = data.productFlows
      expect(data.nav.productFlows, loc).toBeTruthy()
      for (const p of productFlowsCatalog.profiles) {
        expect(pf.profiles[p.id], `${loc}:${p.id}`).toBeTruthy()
        for (const stepId of p.stepIds) {
          expect(pf.steps[stepId], `${loc}:${stepId}`).toBeTruthy()
        }
        for (const fid of p.featureIds) {
          expect(pf.features[fid], `${loc}:${fid}`).toBeTruthy()
        }
      }
      for (const link of productFlowsCatalog.links) {
        expect(pf.links[link.id], `${loc}:${link.id}`).toBeTruthy()
        expect(pf.features[link.featureId], `${loc}:feat:${link.featureId}`).toBeTruthy()
      }
    }
  })
})

describe('parseFlowProfileId', () => {
  it('accepte les IDs connus et refuse le reste', () => {
    expect(parseFlowProfileId('commercial')).toBe('commercial')
    expect(parseFlowProfileId('vet')).toBe('vet')
    expect(parseFlowProfileId('nope')).toBeNull()
    expect(parseFlowProfileId('')).toBeNull()
    expect(parseFlowProfileId(undefined)).toBeNull()
    expect(parseFlowProfileId(['client', 'vet'])).toBe('client')
    expect(parseFlowProfileId(['nope'])).toBeNull()
  })
})


describe('productFlowsNavItem', () => {
  it('construit un item hub', () => {
    expect(productFlowsNavItem('Flux', 'Offre')).toEqual({
      to: '/flux',
      label: 'Flux',
      icon: 'hub',
      section: 'Offre',
    })
  })
})
