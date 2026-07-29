import { describe, expect, it } from 'vitest'
import { isFoodChainSpecies, PET_SPECIES_CODES } from '../../utils/pet-species'

describe('pet-species', () => {
  it('lists production and companion species', () => {
    expect(PET_SPECIES_CODES).toContain('cattle')
    expect(PET_SPECIES_CODES).toContain('alpaca')
    expect(PET_SPECIES_CODES).toContain('donkey')
    expect(PET_SPECIES_CODES.at(-1)).toBe('other')
  })

  it('flags food-chain species for regulatory UI', () => {
    expect(isFoodChainSpecies('cattle')).toBe(true)
    expect(isFoodChainSpecies('horse')).toBe(true)
    expect(isFoodChainSpecies('alpaca')).toBe(true)
    expect(isFoodChainSpecies('dog')).toBe(false)
    expect(isFoodChainSpecies('other')).toBe(false)
    expect(isFoodChainSpecies(undefined)).toBe(false)
  })
})
