import { describe, expect, it } from 'vitest'
import fr from '../../locales/fr.json'
import en from '../../locales/en.json'
import nl from '../../locales/nl.json'
import es from '../../locales/es.json'
import et from '../../locales/et.json'
import itLocale from '../../locales/it.json'

const locales = { fr, en, nl, es, et, it: itLocale } as const

const requiredWalkinKeys = [
  'badge',
  'identifyTitle',
  'identifyHint',
  'modeCreate',
  'modeExisting',
  'emailOptional',
  'petName',
  'petSpecies',
  'confirmExisting',
  'createNewPet',
  'submit',
  'readonlyHint',
  'callbackPhone',
  'callbackPhonePlaceholder',
  'callbackPhoneHint',
] as const

describe('clients.walkin i18n', () => {
  for (const [code, catalog] of Object.entries(locales)) {
    it(`${code} has clients.walkin keys`, () => {
      const walkin = (catalog as any).clients?.walkin
      expect(walkin, `${code} clients.walkin`).toBeTruthy()
      for (const key of requiredWalkinKeys) {
        expect(typeof walkin[key], `${code}.${key}`).toBe('string')
        expect(String(walkin[key]).length).toBeGreaterThan(0)
      }
    })
  }
})
