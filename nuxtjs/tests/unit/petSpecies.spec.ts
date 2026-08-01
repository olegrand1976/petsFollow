import { describe, expect, it } from 'vitest'
import {
  foodChainToYesNo,
  isFoodChainSpecies,
  PET_SPECIES_CODES,
  yesNoToFoodChain,
} from '../../utils/pet-species'

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

  it('maps food-chain yes/no without wiping excluded', () => {
    expect(foodChainToYesNo('food_producing')).toBe('yes')
    expect(foodChainToYesNo('companion')).toBe('no')
    expect(foodChainToYesNo('excluded_from_food_chain')).toBe('no')
    expect(yesNoToFoodChain('yes')).toBe('food_producing')
    expect(yesNoToFoodChain('no', 'companion')).toBe('companion')
    expect(yesNoToFoodChain('no', 'excluded_from_food_chain')).toBe('excluded_from_food_chain')
    expect(yesNoToFoodChain('no', 'food_producing')).toBe('companion')
  })
})

