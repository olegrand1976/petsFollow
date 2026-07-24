import { describe, expect, it, vi, beforeEach } from 'vitest'
import {
  extractAccessToken,
  isAuthSuccess,
  isMFAChallenge,
  isProRole,
  isSalesForceRole,
  hasSessionCookie,
  homePathForRole,
  parseJwtRole,
  unwrapAuthData,
  clearAuthTokens,
  type AuthMFAChallenge,
  type AuthTokens,
} from '../../composables/useAuth'

const cookieStore = new Map<string, string | null>()
const stateStore = new Map<string, { value: unknown }>()

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

vi.stubGlobal('useState', (key: string, init?: () => unknown) => {
  if (!stateStore.has(key)) {
    stateStore.set(key, { value: init ? init() : null })
  }
  return stateStore.get(key)!
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
    stateStore.clear()
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

  it('isAuthSuccess accepte le flag BFF ou les tokens legacy, refuse le MFA', () => {
    expect(isAuthSuccess({ authenticated: true })).toBe(true)
    expect(isAuthSuccess(tokens)).toBe(true)
    expect(isAuthSuccess({})).toBe(false)
    expect(isAuthSuccess(mfa)).toBe(false)
  })

  it('hasSessionCookie détecte pf_token / pf_refresh / pf_session', () => {
    expect(hasSessionCookie()).toBe(false)
    cookieStore.set('pf_session', '1')
    expect(hasSessionCookie()).toBe(true)
    cookieStore.set('pf_session', null)
    cookieStore.set('pf_token', 'jwt')
    expect(hasSessionCookie()).toBe(true)
  })

  it('clearAuthTokens purge les cookies visibles et le profil Pro', async () => {
    cookieStore.set('pf_token', 'access.jwt')
    cookieStore.set('pf_refresh', 'refresh.jwt')
    cookieStore.set('pf_session', '1')
    useState('pro-user').value = { role: 'vet' }
    await clearAuthTokens()
    expect(cookieStore.get('pf_token')).toBeNull()
    expect(cookieStore.get('pf_refresh')).toBeNull()
    expect(cookieStore.get('pf_session')).toBeNull()
    expect(useState('pro-user').value).toBeNull()
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
