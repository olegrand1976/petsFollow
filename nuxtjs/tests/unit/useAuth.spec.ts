import { describe, expect, it, vi, beforeEach } from 'vitest'
import {
  extractAccessToken,
  isAuthSuccess,
  isMFAChallenge,
  isPracticeStaffRole,
  isProRole,
  isSalesForceRole,
  hasSessionCookie,
  homePathForRole,
  parseJwtRole,
  unwrapAuthData,
  clearAuthTokens,
  markAuthSessionActive,
  finishClientLoginSession,
  AUTH_POST_LOGIN_RELOAD_PATH,
  AUTH_LOGIN_REASON_PRO_ONLY,
  AUTH_POST_LOGIN_GRACE_MS,
  isWithinPostLoginGrace,
  clearAuthTokensUnlessPostLoginGrace,
  resolvePostLoginTarget,
  fetchUserAfterLogin,
  authErrorStatus,
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

  function jwtWithPayload(payload: Record<string, unknown>) {
    const body = btoa(JSON.stringify(payload))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '')
    return `hdr.${body}.sig`
  }

  it('parseJwtRole décode le rôle du payload', () => {
    expect(parseJwtRole(jwtWithPayload({ role: 'vet', sub: 'u1' }))).toBe('vet')
    expect(parseJwtRole('bad')).toBeNull()
    expect(parseJwtRole(null)).toBeNull()
  })

  it('parseJwtRole ignore un JWT expiré', () => {
    expect(parseJwtRole(jwtWithPayload({ role: 'vet', exp: 1 }))).toBeNull()
    const futureExp = Math.floor(Date.now() / 1000) + 3600
    expect(parseJwtRole(jwtWithPayload({ role: 'vet', exp: futureExp }))).toBe('vet')
  })

  it('parseJwtRole reconnaît commercial', () => {
    expect(parseJwtRole(jwtWithPayload({ role: 'commercial', sub: 'c1' }))).toBe('commercial')
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

  it('finishClientLoginSession arme la grâce et replace vers un path interne', () => {
    const replace = vi.fn()
    const prevWindow = globalThis.window
    vi.stubGlobal('window', { location: { replace } })
    const store = new Map<string, string>()
    store.set('pf_desk_locked', '1')
    vi.stubGlobal('sessionStorage', {
      getItem: (k: string) => store.get(k) ?? null,
      setItem: (k: string, v: string) => { store.set(k, v) },
      removeItem: (k: string) => { store.delete(k) },
    })
    try {
      finishClientLoginSession()
      expect(isWithinPostLoginGrace()).toBe(true)
      expect(store.has('pf_desk_locked')).toBe(false)
      expect(replace).toHaveBeenCalledWith(AUTH_POST_LOGIN_RELOAD_PATH)
      finishClientLoginSession('/dashboard')
      expect(replace).toHaveBeenCalledWith('/dashboard')
      // open-redirect refusé
      finishClientLoginSession('https://evil.example/')
      expect(replace).toHaveBeenCalledWith(AUTH_POST_LOGIN_RELOAD_PATH)
      finishClientLoginSession('//evil.example/')
      expect(replace).toHaveBeenLastCalledWith(AUTH_POST_LOGIN_RELOAD_PATH)
    } finally {
      if (prevWindow === undefined) {
        // @ts-expect-error restore absent window in node
        delete globalThis.window
      } else {
        vi.stubGlobal('window', prevWindow)
      }
    }
  })

  it('AUTH_LOGIN_REASON_PRO_ONLY est stable pour /login?reason=', () => {
    expect(AUTH_LOGIN_REASON_PRO_ONLY).toBe('proOnly')
  })

  it('post-login grace bloque clearAuthTokens pendant AUTH_POST_LOGIN_GRACE_MS', async () => {
    const store = new Map<string, string>()
    vi.stubGlobal('sessionStorage', {
      getItem: (k: string) => store.get(k) ?? null,
      setItem: (k: string, v: string) => { store.set(k, v) },
      removeItem: (k: string) => { store.delete(k) },
    })
    expect(isWithinPostLoginGrace()).toBe(false)
    markAuthSessionActive()
    expect(isWithinPostLoginGrace()).toBe(true)
    expect(AUTH_POST_LOGIN_GRACE_MS).toBeGreaterThan(1000)
    cookieStore.set('pf_session', '1')
    const cleared = await clearAuthTokensUnlessPostLoginGrace()
    expect(cleared).toBe(false)
    expect(cookieStore.get('pf_session')).toBe('1')
    // Expire la grâce
    store.set('pf_auth_grace_until', String(Date.now() - 1))
    expect(isWithinPostLoginGrace()).toBe(false)
  })

  it('clearAuthTokens purge pf_session et le profil Pro (JWT via BFF côté client)', async () => {
    cookieStore.set('pf_token', 'access.jwt')
    cookieStore.set('pf_refresh', 'refresh.jwt')
    cookieStore.set('pf_session', '1')
    useState('pro-user').value = { role: 'vet' }
    await clearAuthTokens()
    // Client: httpOnly JWT are left to POST /api/auth/logout — do not touch via useCookie.
    expect(cookieStore.get('pf_token')).toBe('access.jwt')
    expect(cookieStore.get('pf_refresh')).toBe('refresh.jwt')
    expect(cookieStore.get('pf_session')).toBeNull()
    expect(useState('pro-user').value).toBeNull()
    // Stale request cookies must not re-arm hasSession (SSR redirect loop).
    expect(hasSessionCookie()).toBe(false)
    markAuthSessionActive()
    expect(hasSessionCookie()).toBe(true)
  })

  it('isProRole / isSalesForceRole couvrent les rôles Pro', () => {
    expect(isProRole('admin')).toBe(true)
    expect(isProRole('vet')).toBe(true)
    expect(isProRole('vet_assistant')).toBe(true)
    expect(isProRole('secretary')).toBe(true)
    expect(isProRole('commercial')).toBe(true)
    expect(isProRole('commercial_manager')).toBe(true)
    expect(isProRole('client')).toBe(false)
    expect(isPracticeStaffRole('vet')).toBe(true)
    expect(isPracticeStaffRole('vet_assistant')).toBe(true)
    expect(isPracticeStaffRole('secretary')).toBe(true)
    expect(isPracticeStaffRole('admin')).toBe(false)
    expect(isSalesForceRole('commercial')).toBe(true)
    expect(isSalesForceRole('commercial_manager')).toBe(true)
    expect(isSalesForceRole('vet')).toBe(false)
  })

  it('homePathForRole route chaque rôle Pro', () => {
    expect(homePathForRole('admin')).toBe('/admin')
    expect(homePathForRole('commercial')).toBe('/commercial')
    expect(homePathForRole('commercial_manager')).toBe('/commercial-manager')
    expect(homePathForRole('vet')).toBe('/dashboard')
    expect(homePathForRole('vet_assistant')).toBe('/dashboard')
    expect(homePathForRole('secretary')).toBe('/dashboard')
    expect(homePathForRole('vet', { profileComplete: false })).toBe('/onboarding')
    expect(homePathForRole('vet', { profileComplete: true })).toBe('/dashboard')
    expect(homePathForRole('client')).toBe('/login')
    expect(homePathForRole(null)).toBe('/login')
  })

  it('parseJwtRole reconnaît commercial_manager', () => {
    expect(parseJwtRole(jwtWithPayload({ role: 'commercial_manager', sub: 'm1' }))).toBe(
      'commercial_manager',
    )
  })

  describe('resolvePostLoginTarget', () => {
    it('ne confond pas /me null avec proOnly (login OK + cookies absents)', () => {
      expect(resolvePostLoginTarget(null)).toEqual({ kind: 'sessionUnavailable' })
      expect(resolvePostLoginTarget(undefined)).toEqual({ kind: 'sessionUnavailable' })
      expect(resolvePostLoginTarget({})).toEqual({ kind: 'sessionUnavailable' })
    })

    it('redirige commercial / manager vers leur home', () => {
      expect(resolvePostLoginTarget({ role: 'commercial', contactPhone: '0470' })).toEqual({
        kind: 'navigate',
        path: '/commercial',
      })
      expect(resolvePostLoginTarget({ role: 'commercial_manager', contactPhone: '0470' })).toEqual({
        kind: 'navigate',
        path: '/commercial-manager',
      })
    })

    it('force complete-contact-phone si commercial sans téléphone', () => {
      expect(resolvePostLoginTarget({ role: 'commercial', contactPhone: '' })).toEqual({
        kind: 'navigate',
        path: '/complete-contact-phone',
      })
    })

    it('force change-password avant le home', () => {
      expect(resolvePostLoginTarget({
        role: 'commercial',
        mustChangePassword: true,
      })).toEqual({ kind: 'navigate', path: '/change-password' })
    })

    it('refuse un rôle client (proOnly)', () => {
      expect(resolvePostLoginTarget({ role: 'client' })).toEqual({ kind: 'proOnly' })
    })

    it('accepte un rôle JWT si /me soft-null (fallback legacy/SSR)', () => {
      expect(resolvePostLoginTarget(null, 'commercial')).toEqual({
        kind: 'navigate',
        path: '/commercial',
      })
    })
  })

  describe('fetchUserAfterLogin', () => {
    it('retry sur soft-null puis retourne le profil', async () => {
      const sleep = vi.fn(async () => {})
      const fetchUser = vi
        .fn()
        .mockResolvedValueOnce(null)
        .mockResolvedValueOnce({ role: 'commercial' })

      const me = await fetchUserAfterLogin(fetchUser, { attempts: 3, delayMs: 10, sleep })
      expect(me).toEqual({ role: 'commercial' })
      expect(fetchUser).toHaveBeenCalledTimes(2)
      expect(sleep).toHaveBeenCalledTimes(1)
    })

    it('retry sur 401 puis propage si toujours KO', async () => {
      const sleep = vi.fn(async () => {})
      const err = { statusCode: 401 }
      const fetchUser = vi.fn().mockRejectedValue(err)

      await expect(
        fetchUserAfterLogin(fetchUser, { attempts: 3, delayMs: 10, sleep }),
      ).rejects.toEqual(err)
      expect(fetchUser).toHaveBeenCalledTimes(3)
      expect(sleep).toHaveBeenCalledTimes(2)
    })

    it('ne confond plus soft-null final avec un throw (login OK + /me échec)', async () => {
      const sleep = vi.fn(async () => {})
      const fetchUser = vi.fn().mockResolvedValue(null)

      await expect(
        fetchUserAfterLogin(fetchUser, { attempts: 2, delayMs: 5, sleep }),
      ).resolves.toBeNull()
      expect(resolvePostLoginTarget(null)).toEqual({ kind: 'sessionUnavailable' })
    })

    it('récupère après un 401 transitoire (course Set-Cookie)', async () => {
      const sleep = vi.fn(async () => {})
      const fetchUser = vi
        .fn()
        .mockRejectedValueOnce({ statusCode: 401 })
        .mockResolvedValueOnce({ role: 'commercial', profileComplete: true })

      const me = await fetchUserAfterLogin(fetchUser, { attempts: 3, delayMs: 10, sleep })
      expect(me?.role).toBe('commercial')
      expect(fetchUser).toHaveBeenCalledTimes(2)
    })
  })

  describe('authErrorStatus', () => {
    it('lit statusCode / status / response.status', () => {
      expect(authErrorStatus({ statusCode: 401 })).toBe(401)
      expect(authErrorStatus({ status: 403 })).toBe(403)
      expect(authErrorStatus({ response: { status: 500 } })).toBe(500)
      expect(authErrorStatus({})).toBeNull()
    })
  })
})
