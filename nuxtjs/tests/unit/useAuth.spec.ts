import { describe, expect, it } from 'vitest'
import {
  extractAccessToken,
  isMFAChallenge,
  parseJwtRole,
  unwrapAuthData,
  type AuthMFAChallenge,
  type AuthTokens,
} from '../../composables/useAuth'

describe('useAuth helpers', () => {
  const tokens: AuthTokens = {
    accessToken: 'access.jwt',
    refreshToken: 'refresh.jwt',
    expiresIn: 900,
  }

  const mfa: AuthMFAChallenge = {
    requires2FA: true,
    mfaToken: 'mfa.jwt',
    expiresIn: 300,
  }

  it('unwrapAuthData lit data enveloppé ou brut', () => {
    expect(unwrapAuthData({ data: tokens })).toEqual(tokens)
    expect(unwrapAuthData(tokens)).toEqual(tokens)
  })

  it('isMFAChallenge distingue MFA et tokens', () => {
    expect(isMFAChallenge(mfa)).toBe(true)
    expect(isMFAChallenge(tokens)).toBe(false)
  })

  it('extractAccessToken ignore le challenge MFA', () => {
    expect(extractAccessToken(tokens)).toBe('access.jwt')
    expect(extractAccessToken(mfa)).toBeNull()
  })

  it('parseJwtRole décode le rôle du payload', () => {
    const payload = btoa(JSON.stringify({ role: 'vet', sub: 'u1' }))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '')
    const token = `hdr.${payload}.sig`
    expect(parseJwtRole(token)).toBe('vet')
    expect(parseJwtRole('bad')).toBeNull()
    expect(parseJwtRole(null)).toBeNull()
  })
})
