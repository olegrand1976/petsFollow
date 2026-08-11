import { describe, expect, it } from 'vitest'
import { buildCsp } from '../../utils/buildCsp'

describe('buildCsp', () => {
  it('conserve localhost explicite (CI Playwright / WS pitch)', () => {
    const csp = buildCsp('http://localhost:8291')
    expect(csp).toContain('http://localhost:8291')
    expect(csp).toContain('ws://localhost:8291')
    expect(csp).toMatch(/media-src 'self' blob: http:\/\/localhost:8291/)
  })

  it('inclut l\'API publique https dans connect-src et media-src', () => {
    const csp = buildCsp('https://api.petsfollow.ll-it-sc.be')
    expect(csp).toContain('https://api.petsfollow.ll-it-sc.be')
    expect(csp).toContain('wss://api.petsfollow.ll-it-sc.be')
    expect(csp).toMatch(/media-src 'self' blob: https:\/\/api\.petsfollow\.ll-it-sc\.be/)
  })

  it('sans env : self + Google uniquement (pas de localhost inventé)', () => {
    const csp = buildCsp('')
    expect(csp).toContain("connect-src 'self' https://accounts.google.com/gsi/")
    expect(csp).not.toContain('localhost')
    expect(csp).not.toContain('127.0.0.1')
  })

  it('autorise wasm Cornerstone (PACS) sans élargir script-src à *', () => {
    const csp = buildCsp('')
    expect(csp).toContain("'wasm-unsafe-eval'")
    expect(csp).toMatch(/script-src[^;]*'self'/)
    expect(csp).not.toMatch(/script-src[^;]*\*/)
    expect(csp).toContain("worker-src 'self' blob:")
  })

  it('autorise iframe same-origin (viewer PDF admin) + GSI', () => {
    const csp = buildCsp('')
    expect(csp).toMatch(/frame-src 'self' https:\/\/accounts\.google\.com\/gsi\//)
    expect(csp).toContain("frame-ancestors 'none'")
  })

  it('autorise frame-ancestors self pour réponses PDF iframe', () => {
    const csp = buildCsp('', { frameAncestors: "'self'" })
    expect(csp).toContain("frame-ancestors 'self'")
    expect(csp).not.toContain("frame-ancestors 'none'")
  })
})
