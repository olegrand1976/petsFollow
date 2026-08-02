import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const localesDir = join(__dirname, '../../locales')
const locales = ['fr', 'en', 'nl', 'es', 'et', 'it'] as const

const expectedAbbr: Record<(typeof locales)[number], string> = {
  fr: 'FR',
  en: 'RR',
  nl: 'AF',
  es: 'FR',
  et: 'RR',
  it: 'FR',
}

const expectedFull: Record<(typeof locales)[number], string> = {
  fr: 'Fréquence respiratoire',
  en: 'Respiratory rate',
  nl: 'Ademhalingsfrequentie',
  es: 'Frecuencia respiratoria',
  et: 'Hingamissagedus',
  it: 'Frequenza respiratoria',
}

describe('pets reading type i18n (FR / respiratory)', () => {
  for (const loc of locales) {
    it(`${loc}: abbr + full label for heartrate reading type`, () => {
      const json = JSON.parse(readFileSync(join(localesDir, `${loc}.json`), 'utf8')) as {
        pets: {
          readingType: { heartrate: string }
          readingTypeFull: { heartrate: string }
          columnLastHeartRate: string
        }
      }
      expect(json.pets.readingType.heartrate).toBe(expectedAbbr[loc])
      expect(json.pets.readingTypeFull.heartrate).toBe(expectedFull[loc])
      expect(json.pets.columnLastHeartRate).toMatch(/FR|RR|AF/)
      expect(json.pets.columnLastHeartRate).not.toMatch(/\bFC\b|\bHR\b|\bHF\b/)
    })
  }
})
