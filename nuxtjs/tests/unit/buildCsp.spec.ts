import { describe, expect, it } from 'vitest'
import { buildCsp } from '../../utils/buildCsp'

describe('buildCsp', () => {
  it('n\'inclut pas localhost / loopback dans connect-src', () => {
    const csp = buildCsp('http://localhost:8291')
    expect(csp).toContain("connect-src 'self' https://accounts.google.com/gsi/")
    expect(csp).not.toContain('localhost')
    expect(csp).not.toContain('127.0.0.1')
    expect(csp).toContain("media-src 'self' blob:")
    expect(csp).not.toMatch(/media-src[^;]*localhost/)
  })

  it('inclut l\'API publique https dans connect-src et media-src', () => {
    const csp = buildCsp('https://api.petsfollow.ll-it-sc.be')
    expect(csp).toContain('https://api.petsfollow.ll-it-sc.be')
    expect(csp).toContain('wss://api.petsfollow.ll-it-sc.be')
    expect(csp).toMatch(/media-src 'self' blob: https:\/\/api\.petsfollow\.ll-it-sc\.be/)
  })

  it('sans env : self + Google uniquement', () => {
    const csp = buildCsp('')
    expect(csp).toContain("connect-src 'self' https://accounts.google.com/gsi/")
    expect(csp).not.toContain('localhost')
  })
})
