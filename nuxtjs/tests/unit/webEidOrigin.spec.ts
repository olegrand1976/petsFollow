import { beforeEach, describe, expect, it, vi } from 'vitest'

/**
 * GET same-origin n’envoie souvent pas Origin — la BFF doit dériver l’origine
 * onglet (Host / Referer) et forcer X-PF-Proxy-Secret pour que Go accepte
 * X-PF-Web-Eid-Origin (alias localhost ↔ 127.0.0.1).
 */

let headers: Record<string, string> = {}
let requestURL = new URL('http://127.0.0.1:3002/api/vet/eid/web-eid/challenge')
let bffProxySecret = 'dev-bff-proxy-secret'

vi.stubGlobal('useRuntimeConfig', () => ({
  apiBase: 'http://api.test',
  bffProxySecret,
}))
vi.stubGlobal('getCookie', () => undefined)
vi.stubGlobal('getRequestHeader', (_event: unknown, name: string) => headers[name])
vi.stubGlobal('getRequestIP', () => undefined)
vi.stubGlobal('getRequestURL', () => requestURL)

const { resolveWebEidOrigin, webEidUpstreamHeaders } = await import('../../server/utils/api')

describe('resolveWebEidOrigin', () => {
  beforeEach(() => {
    headers = {}
    requestURL = new URL('http://127.0.0.1:3002/api/vet/eid/web-eid/challenge')
    bffProxySecret = 'dev-bff-proxy-secret'
    vi.stubGlobal('getRequestURL', () => requestURL)
  })

  it('préfère le header Origin', () => {
    headers = { origin: 'http://localhost:3002/' }
    expect(resolveWebEidOrigin({} as any)).toBe('http://localhost:3002')
  })

  it('sans Origin, dérive getRequestURL (cas GET same-origin 127.0.0.1)', () => {
    headers = {}
    expect(resolveWebEidOrigin({} as any)).toBe('http://127.0.0.1:3002')
  })

  it('sans Origin ni URL, retombe sur Referer', () => {
    headers = { referer: 'http://127.0.0.1:3002/clients/new' }
    vi.stubGlobal('getRequestURL', () => {
      throw new Error('no url')
    })
    expect(resolveWebEidOrigin({} as any)).toBe('http://127.0.0.1:3002')
  })

  it('rejette une origine non http(s)', () => {
    headers = { origin: 'ftp://localhost:3002' }
    vi.stubGlobal('getRequestURL', () => {
      throw new Error('no url')
    })
    expect(resolveWebEidOrigin({} as any)).toBe('')
  })
})

describe('webEidUpstreamHeaders', () => {
  beforeEach(() => {
    headers = {}
    requestURL = new URL('http://127.0.0.1:3002/api/vet/eid/web-eid/challenge')
    bffProxySecret = 'dev-bff-proxy-secret'
    vi.stubGlobal('getRequestURL', () => requestURL)
  })

  it('pose origine + secret même sans IP client', () => {
    const h = webEidUpstreamHeaders({} as any)
    expect(h['X-PF-Web-Eid-Origin']).toBe('http://127.0.0.1:3002')
    expect(h['X-PF-Proxy-Secret']).toBe('dev-bff-proxy-secret')
  })

  it('pose l’origine même sans secret (Go ignore sans X-PF-Proxy-Secret)', () => {
    bffProxySecret = ''
    const h = webEidUpstreamHeaders({} as any)
    expect(h['X-PF-Web-Eid-Origin']).toBe('http://127.0.0.1:3002')
    expect(h['X-PF-Proxy-Secret']).toBeUndefined()
  })
})
