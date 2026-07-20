import { describe, expect, it, vi, beforeEach } from 'vitest'
import {
  extractAccessToken,
  isMFAChallenge,
  isProRole,
  isSalesForceRole,
  homePathForRole,
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

  it('parseJwtRole reconnaît commercial', () => {
    const payload = btoa(JSON.stringify({ role: 'commercial', sub: 'c1' }))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '')
    expect(parseJwtRole(`hdr.${payload}.sig`)).toBe('commercial')
  })

  it('persistAuthTokens / clearAuthTokens gèrent pf_token et pf_refresh', () => {
    persistAuthTokens(tokens)
    expect(cookieStore.get('pf_token')).toBe('access.jwt')
    expect(cookieStore.get('pf_refresh')).toBe('refresh.jwt')
    clearAuthTokens()
    expect(cookieStore.get('pf_token')).toBeNull()
    expect(cookieStore.get('pf_refresh')).toBeNull()
  })

  it('isProRole / isSalesForceRole couvrent les rôles Pro', () => {
    expect(isProRole('admin')).toBe(true)
    expect(isProRole('vet')).toBe(true)
    expect(isProRole('commercial')).toBe(true)
    expect(isProRole('commercial_manager')).toBe(true)
    expect(isProRole('client')).toBe(false)
    expect(isSalesForceRole('commercial')).toBe(true)
    expect(isSalesForceRole('commercial_manager')).toBe(true)
    expect(isSalesForceRole('vet')).toBe(false)
  })

  it('homePathForRole route chaque rôle Pro', () => {
    expect(homePathForRole('admin')).toBe('/admin')
    expect(homePathForRole('commercial')).toBe('/commercial')
    expect(homePathForRole('commercial_manager')).toBe('/commercial-manager')
    expect(homePathForRole('vet')).toBe('/dashboard')
    expect(homePathForRole('vet', { profileComplete: false })).toBe('/onboarding')
    expect(homePathForRole('vet', { profileComplete: true })).toBe('/dashboard')
    expect(homePathForRole('client')).toBe('/login')
    expect(homePathForRole(null)).toBe('/login')
  })

  it('parseJwtRole reconnaît commercial_manager', () => {
    const payload = btoa(JSON.stringify({ role: 'commercial_manager', sub: 'm1' }))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '')
    expect(parseJwtRole(`hdr.${payload}.sig`)).toBe('commercial_manager')
  })
})
