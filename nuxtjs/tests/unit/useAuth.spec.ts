import { describe, expect, it, vi, beforeEach } from 'vitest'
import {
  extractAccessToken,
  isMFAChallenge,
  parseJwtRole,
  unwrapAuthData,
  persistAuthTokens,
  clearAuthTokens,
  type AuthMFAChallenge,
  type AuthTokens,
} from '../../composables/useAuth'

const cookieStore = new Map<string, string | null>()

vi.stubGlobal('useCookie', (name: string, _opts?: unknown) => {
  if (!cookieStore.has(name)) cookieStore.set(name, null)
  return {
    get value() {
      return cookieStore.get(name) ?? null
    },
    set value(v: string | null) {
      cookieStore.set(name, v)
    },
  }
})

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

  beforeEach(() => {
    cookieStore.clear()
  })

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

  it('persistAuthTokens / clearAuthTokens gèrent pf_token et pf_refresh', () => {
    persistAuthTokens(tokens)
    expect(cookieStore.get('pf_token')).toBe('access.jwt')
    expect(cookieStore.get('pf_refresh')).toBe('refresh.jwt')
    clearAuthTokens()
    expect(cookieStore.get('pf_token')).toBeNull()
    expect(cookieStore.get('pf_refresh')).toBeNull()
  })
})
