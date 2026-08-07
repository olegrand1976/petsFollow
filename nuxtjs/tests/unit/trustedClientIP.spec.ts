import { describe, expect, it, vi } from 'vitest'

/**
 * La BFF relaie l'IP appelante à l'API Go via X-PF-Client-IP + secret partagé.
 * Sans secret, aucun header d'IP n'est posé (l'API rate-limite sur l'edge).
 */

let headers: Record<string, string> = {}
let socketIP: string | undefined = '10.0.0.1'
let bffProxySecret = 'dev-bff-proxy-secret'

vi.stubGlobal('useRuntimeConfig', () => ({
  apiBase: 'http://api.test',
  bffProxySecret,
}))
vi.stubGlobal('getCookie', () => undefined)
vi.stubGlobal('getRequestHeader', (_event: unknown, name: string) => headers[name])
vi.stubGlobal('getRequestIP', () => socketIP)

const { trustedClientIP, localeHeaders } = await import('../../server/utils/api')

describe('trustedClientIP', () => {
  it('retient la dernière entrée de X-Forwarded-For, posée par notre hop', () => {
    headers = { 'x-forwarded-for': '203.0.113.7, 198.51.100.4' }
    expect(trustedClientIP({} as any)).toBe('198.51.100.4')
  })

  it('ignore une valeur forgée par le navigateur', () => {
    headers = { 'x-forwarded-for': '1.1.1.1, 198.51.100.4' }
    expect(trustedClientIP({} as any)).toBe('198.51.100.4')
  })

  it('rejette une entrée qui n’est pas une IP', () => {
    headers = { 'x-forwarded-for': 'not-an-ip' }
    socketIP = undefined
    expect(trustedClientIP({} as any)).toBeUndefined()
    socketIP = '10.0.0.1'
  })

  it('retombe sur l’IP de socket sans X-Forwarded-For', () => {
    headers = {}
    socketIP = '10.0.0.1'
    expect(trustedClientIP({} as any)).toBe('10.0.0.1')
  })
})

describe('localeHeaders BFF→API', () => {
  it('pose X-PF-Client-IP + secret, pas X-Forwarded-For', () => {
    headers = { 'x-forwarded-for': '203.0.113.7, 198.51.100.4' }
    bffProxySecret = 'dev-bff-proxy-secret'
    const h = localeHeaders({} as any)
    expect(h['X-PF-Client-IP']).toBe('198.51.100.4')
    expect(h['X-PF-Proxy-Secret']).toBe('dev-bff-proxy-secret')
    expect(h['X-Forwarded-For']).toBeUndefined()
  })

  it('ne pose aucun header d’IP sans secret (fail-closed)', () => {
    headers = { 'x-forwarded-for': '198.51.100.4' }
    bffProxySecret = ''
    const h = localeHeaders({} as any)
    expect(h['X-PF-Client-IP']).toBeUndefined()
    expect(h['X-PF-Proxy-Secret']).toBeUndefined()
    expect(h['X-Forwarded-For']).toBeUndefined()
    bffProxySecret = 'dev-bff-proxy-secret'
  })
})
