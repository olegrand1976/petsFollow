export type VetPetListRow = {
  id: string
  name: string
  species: string
  breed?: string
  ownerName: string
  lastHeartRateAt?: string
  unreadHeartrateCount?: number
}

export type VetPetReadingType = 'heartrate'

export type VetPetsListFilters = {
  query?: string
  species?: string
  unreadOnly?: boolean
}

/** Reading type for the practice pets list (FR / respiratory; technical key heartrate). */
export function petReadingType(pet: Pick<VetPetListRow, 'lastHeartRateAt' | 'unreadHeartrateCount'>): VetPetReadingType | null {
  if (pet.lastHeartRateAt || (pet.unreadHeartrateCount ?? 0) > 0) return 'heartrate'
  return null
}

export function filterVetPetsList<T extends VetPetListRow>(pets: T[], filters: VetPetsListFilters = {}): T[] {
  const q = (filters.query ?? '').trim().toLowerCase()
  const species = filters.species ?? 'all'
  const unreadOnly = Boolean(filters.unreadOnly)
  return pets.filter((p) => {
    if (species !== 'all' && p.species !== species) return false
    if (unreadOnly && (p.unreadHeartrateCount ?? 0) <= 0) return false
    if (!q) return true
    const hay = [p.name, p.breed, p.ownerName, p.species].filter(Boolean).join(' ').toLowerCase()
    return hay.includes(q)
  })
}
