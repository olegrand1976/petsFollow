import { describe, expect, it } from 'vitest'
import catalog from '../../data/usecases/catalog.json'
import { parseAppEnv } from '../../composables/useAppEnv'
import { usecasesBasePath, usecasesNavItem } from '../../utils/usecasesPath'

describe('usecases catalog', () => {
  it('contient l’arborescence et les IDs démo', () => {
    expect(catalog.folders.length).toBeGreaterThanOrEqual(8)
    expect(Object.keys(catalog.cases).length).toBeGreaterThanOrEqual(19)
    for (const id of catalog.demoSession) {
      expect(catalog.cases[id], id).toBeTruthy()
      expect(catalog.cases[id].steps.length).toBeGreaterThan(0)
    }
  })
})

describe('usecasesBasePath', () => {
  it('pointe vers /usecases pour tous les rôles', () => {
    expect(usecasesBasePath('admin')).toBe('/usecases')
    expect(usecasesBasePath('commercial')).toBe('/usecases')
    expect(usecasesBasePath('commercial_manager')).toBe('/usecases')
  })
})

describe('usecasesNavItem', () => {
  it('construit un item checklist', () => {
    expect(usecasesNavItem('Guides', 'Offre')).toEqual({
      to: '/usecases',
      label: 'Guides',
      icon: 'checklist',
      section: 'Offre',
    })
  })
})

describe('parseAppEnv', () => {
  it('normalise et refuse les valeurs inconnues', () => {
    expect(parseAppEnv('staging')).toBe('staging')
    expect(parseAppEnv('production')).toBe('production')
    expect(parseAppEnv('local')).toBe('local')
    expect(parseAppEnv('weird')).toBe('local')
    expect(parseAppEnv(undefined)).toBe('local')
  })
})
