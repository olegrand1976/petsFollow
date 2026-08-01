/** Canonical pet species codes (aligned with Flutter / Go kernel). */
export const PET_SPECIES_CODES = [
  'dog',
  'cat',
  'horse',
  'donkey',
  'cattle',
  'sheep',
  'goat',
  'pig',
  'poultry',
  'rabbit',
  'alpaca',
  'llama',
  'other',
] as const

export type PetSpeciesCode = (typeof PET_SPECIES_CODES)[number]

/** Species that show domicile / food-chain regulatory UI. */
export function isFoodChainSpecies(species: string | null | undefined): boolean {
  switch (species) {
    case 'horse':
    case 'donkey':
    case 'cattle':
    case 'sheep':
    case 'goat':
    case 'pig':
    case 'poultry':
    case 'rabbit':
    case 'alpaca':
    case 'llama':
      return true
    default:
      return false
  }
}

/** UI oui/non ← API foodChainStatus (excluded counts as « non »). */
export function foodChainToYesNo(status?: string | null): 'yes' | 'no' {
  return status === 'food_producing' ? 'yes' : 'no'
}

/**
 * UI oui/non → API foodChainStatus.
 * Preserves excluded_from_food_chain when the UI stays on « non ».
 */
export function yesNoToFoodChain(v: 'yes' | 'no', current?: string | null): string {
  if (v === 'yes') return 'food_producing'
  if (current === 'excluded_from_food_chain') return 'excluded_from_food_chain'
  return 'companion'
}
