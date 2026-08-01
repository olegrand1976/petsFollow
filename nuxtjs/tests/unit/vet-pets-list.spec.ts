import { describe, expect, it } from 'vitest'
import { filterVetPetsList, petReadingType, type VetPetListRow } from '../../utils/vet-pets-list'

const pets: VetPetListRow[] = [
  {
    id: '1',
    name: 'Rex',
    species: 'dog',
    breed: 'Labrador',
    ownerName: 'Alice',
    lastHeartRateAt: '2026-08-01T10:00:00Z',
    unreadHeartrateCount: 2,
  },
  {
    id: '2',
    name: 'Mimi',
    species: 'cat',
    ownerName: 'Bob',
    lastHeartRateAt: '2026-07-01T10:00:00Z',
    unreadHeartrateCount: 0,
  },
  {
    id: '3',
    name: 'Spirit',
    species: 'horse',
    ownerName: 'Alice',
    unreadHeartrateCount: 0,
  },
]

describe('vet-pets-list', () => {
  it('petReadingType is heartrate when last FC or unread exists', () => {
    expect(petReadingType(pets[0])).toBe('heartrate')
    expect(petReadingType(pets[1])).toBe('heartrate')
    expect(petReadingType(pets[2])).toBeNull()
    expect(petReadingType({ unreadHeartrateCount: 1 })).toBe('heartrate')
  })

  it('filters unread only', () => {
    expect(filterVetPetsList(pets, { unreadOnly: true }).map((p) => p.id)).toEqual(['1'])
  })

  it('combines species + search + unread', () => {
    expect(
      filterVetPetsList(pets, { species: 'dog', query: 'alice', unreadOnly: true }).map((p) => p.id),
    ).toEqual(['1'])
    expect(filterVetPetsList(pets, { species: 'cat', unreadOnly: true })).toEqual([])
  })
})
