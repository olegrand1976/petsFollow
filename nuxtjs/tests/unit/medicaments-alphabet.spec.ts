import { describe, expect, it } from 'vitest'
import {
  MEDICAMENT_ALPHABET_LETTERS,
  afmpsSourceKind,
  letterHasEntries,
  normalizeMedicamentLetter,
  parseAfmpsMeta,
  parseLetterCounts,
} from '~/utils/medicaments-alphabet'

describe('medicaments-alphabet', () => {
  it('exposes A–Z plus hash bucket', () => {
    expect(MEDICAMENT_ALPHABET_LETTERS).toHaveLength(27)
    expect(MEDICAMENT_ALPHABET_LETTERS[0]).toBe('A')
    expect(MEDICAMENT_ALPHABET_LETTERS[25]).toBe('Z')
    expect(MEDICAMENT_ALPHABET_LETTERS[26]).toBe('#')
  })

  it('normalizes letter input', () => {
    expect(normalizeMedicamentLetter('a')).toBe('A')
    expect(normalizeMedicamentLetter(' Z ')).toBe('Z')
    expect(normalizeMedicamentLetter('#')).toBe('#')
    expect(normalizeMedicamentLetter('')).toBeNull()
    expect(normalizeMedicamentLetter('AB')).toBeNull()
    expect(normalizeMedicamentLetter('1')).toBeNull()
  })

  it('fills missing letter counts with zero', () => {
    const counts = parseLetterCounts({ A: 3, V: '2' })
    expect(counts.A).toBe(3)
    expect(counts.V).toBe(2)
    expect(counts.B).toBe(0)
    expect(counts['#']).toBe(0)
    expect(letterHasEntries(counts, 'A')).toBe(true)
    expect(letterHasEntries(counts, 'B')).toBe(false)
  })

  it('parses AFMPS meta snake/camel keys', () => {
    expect(parseAfmpsMeta({
      source: 'afmps-pack-csv',
      manufacturer: 'Lab',
      active_substance: 'amox',
      strength: '50mg',
    })).toEqual({
      source: 'afmps-pack-csv',
      manufacturer: 'Lab',
      activeSubstance: 'amox',
      strength: '50mg',
    })
    expect(parseAfmpsMeta({ activeSubstance: 'x' }).activeSubstance).toBe('x')
    expect(parseAfmpsMeta(null)).toEqual({})
  })

  it('classifies AFMPS source kinds', () => {
    expect(afmpsSourceKind('afmps-pack-csv')).toBe('afmps')
    expect(afmpsSourceKind('afmps_import_cron')).toBe('afmps')
    expect(afmpsSourceKind('AFMPS')).toBe('afmps')
    expect(afmpsSourceKind('compendium-pdf')).toBe('compendium')
    expect(afmpsSourceKind('compendium_pdf')).toBe('compendium')
    expect(afmpsSourceKind('seed')).toBe('other')
    expect(afmpsSourceKind('meta-afmps-leak')).toBe('other')
    expect(afmpsSourceKind(undefined)).toBeNull()
    expect(afmpsSourceKind('')).toBeNull()
  })
})
