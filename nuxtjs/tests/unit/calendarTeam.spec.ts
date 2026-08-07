import { describe, expect, it } from 'vitest'
import {
  extractTeamMembersList,
  mapTeamMembersForCalendar,
} from '../../utils/calendarTeam'

describe('extractTeamMembersList', () => {
  it('unwraps { data: { members } }', () => {
    expect(extractTeamMembersList({ data: { members: [{ userId: 'a' }] } })).toEqual([{ userId: 'a' }])
  })

  it('accepts bare array', () => {
    expect(extractTeamMembersList([{ userId: 'a' }])).toEqual([{ userId: 'a' }])
  })
})

describe('mapTeamMembersForCalendar', () => {
  const rows = [
    { userId: 'u1', fullName: 'Ada', includeInCalendar: true, defaultSiteId: 's1' },
    { userId: 'u2', fullName: 'Bob', includeInCalendar: false, defaultSiteId: '' },
    { userId: 'u3', email: 'c@x', includeInCalendar: true },
    { userId: 'u4', fullName: 'Dan', defaultSiteId: 's2' },
  ]

  it('drops includeInCalendar=false and applies site filter', () => {
    expect(mapTeamMembersForCalendar(rows, 's1').map((m) => m.id)).toEqual(['u1', 'u3'])
  })

  it('treats missing includeInCalendar as true', () => {
    expect(mapTeamMembersForCalendar([{ userId: 'x', fullName: 'X' }])).toEqual([{
      id: 'x',
      fullName: 'X',
      defaultSiteId: '',
      includeInCalendar: true,
    }])
  })

  it('without siteId keeps all included members', () => {
    expect(mapTeamMembersForCalendar(rows).map((m) => m.id)).toEqual(['u1', 'u3', 'u4'])
  })
})
