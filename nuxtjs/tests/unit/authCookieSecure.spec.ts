import { describe, expect, it, afterEach } from 'vitest'
import { authCookieSecure } from '../../utils/authCookieSecure'

describe('authCookieSecure', () => {
  const prevSecure = process.env.AUTH_COOKIE_SECURE
  const prevSite = process.env.PETSFOLLOW_PUBLIC_SITE_URL
  const prevNode = process.env.NODE_ENV

  afterEach(() => {
    if (prevSecure === undefined) delete process.env.AUTH_COOKIE_SECURE
    else process.env.AUTH_COOKIE_SECURE = prevSecure
    if (prevSite === undefined) delete process.env.PETSFOLLOW_PUBLIC_SITE_URL
    else process.env.PETSFOLLOW_PUBLIC_SITE_URL = prevSite
    process.env.NODE_ENV = prevNode
  })

  it('force false via AUTH_COOKIE_SECURE', () => {
    process.env.AUTH_COOKIE_SECURE = '0'
    process.env.NODE_ENV = 'production'
    expect(authCookieSecure({ protocol: 'https:' })).toBe(false)
  })

  it('force true via AUTH_COOKIE_SECURE', () => {
    process.env.AUTH_COOKIE_SECURE = 'true'
    process.env.NODE_ENV = 'development'
    expect(authCookieSecure({ protocol: 'http:' })).toBe(true)
  })

  it('HTTP request → not secure', () => {
    delete process.env.AUTH_COOKIE_SECURE
    process.env.NODE_ENV = 'production'
    expect(authCookieSecure({ protocol: 'http:' })).toBe(false)
  })

  it('HTTPS request → secure', () => {
    delete process.env.AUTH_COOKIE_SECURE
    expect(authCookieSecure({ protocol: 'https:' })).toBe(true)
  })

  it('site URL http → not secure', () => {
    delete process.env.AUTH_COOKIE_SECURE
    process.env.PETSFOLLOW_PUBLIC_SITE_URL = 'http://localhost:3002'
    process.env.NODE_ENV = 'production'
    expect(authCookieSecure()).toBe(false)
  })
})
