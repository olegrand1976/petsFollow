import { beforeEach, describe, expect, it, vi } from 'vitest'

/**
 * Purger les cookies httpOnly ne révoque rien côté serveur : un refresh token
 * qui aurait fuité resterait valide 30 jours. La BFF doit donc appeler
 * POST /auth/logout, qui incrémente token_version.
 */

const cookies: Record<string, string> = {}
const fetchCalls: Array<{ url: string, options: any }> = []
let fetchImpl: (url: string, options: any) => Promise<unknown> = async () => ({})

vi.stubGlobal('useRuntimeConfig', () => ({ apiBase: 'http://api.test' }))
vi.stubGlobal('getCookie', (_event: unknown, name: string) => cookies[name])
vi.stubGlobal('getRequestIP', () => '203.0.113.9')
vi.stubGlobal('$fetch', (url: string, options: any) => {
  fetchCalls.push({ url, options })
  return fetchImpl(url, options)
})

const { revokeIssuedTokens } = await import('../../server/utils/api')

describe('revokeIssuedTokens', () => {
  beforeEach(() => {
    fetchCalls.length = 0
    fetchImpl = async () => ({})
    for (const k of Object.keys(cookies)) delete cookies[k]
    cookies.pf_token = 'access-token-abc'
  })

  it('appelle POST /auth/logout avec le bearer de la session', async () => {
    await revokeIssuedTokens({} as any)

    expect(fetchCalls).toHaveLength(1)
    const [call] = fetchCalls
    expect(call.url).toBe('http://api.test/api/v1/auth/logout')
    expect(call.options.method).toBe('POST')
    expect(call.options.headers.Authorization).toBe('Bearer access-token-abc')
  })

  it("n'échoue pas si l'API est injoignable : la déconnexion locale doit aboutir", async () => {
    fetchImpl = async () => {
      throw new Error('ECONNREFUSED')
    }
    await expect(revokeIssuedTokens({} as any)).resolves.toBeUndefined()
  })

  it('appelle quand même l’API sans cookie de session (token déjà expiré)', async () => {
    delete cookies.pf_token
    await revokeIssuedTokens({} as any)
    expect(fetchCalls).toHaveLength(1)
    expect(fetchCalls[0].options.headers.Authorization).toBeUndefined()
  })
})
