import { describe, expect, it } from 'vitest'
import { roleFromAccessToken } from '../../server/utils/api'

function jwtWithPayload(payload: Record<string, unknown>) {
  const body = Buffer.from(JSON.stringify(payload)).toString('base64url')
  return `hdr.${body}.sig`
}

describe('roleFromAccessToken', () => {
  it('extrait le claim role', () => {
    expect(roleFromAccessToken(jwtWithPayload({ role: 'vet', sub: 'u1' }))).toBe('vet')
    expect(roleFromAccessToken(jwtWithPayload({ role: 'admin' }))).toBe('admin')
  })

  it('ignore token invalide / sans role', () => {
    expect(roleFromAccessToken('not.a.jwt')).toBeUndefined()
    expect(roleFromAccessToken('only-one-part')).toBeUndefined()
    expect(roleFromAccessToken(jwtWithPayload({ sub: 'u1' }))).toBeUndefined()
  })
})
