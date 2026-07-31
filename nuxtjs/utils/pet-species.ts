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
