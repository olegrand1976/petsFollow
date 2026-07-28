import { afterEach, describe, expect, it } from 'vitest'
import { isPublicFlagOn, publicFeatureFlag } from '../../utils/public-feature-flag'

const KEY = 'NUXT_PUBLIC_TEST_FEATURE_FLAG'

describe('publicFeatureFlag', () => {
  const prev = { flag: process.env[KEY], nodeEnv: process.env.NODE_ENV }

  afterEach(() => {
    if (prev.flag === undefined) delete process.env[KEY]
    else process.env[KEY] = prev.flag
    if (prev.nodeEnv === undefined) delete process.env.NODE_ENV
    else process.env.NODE_ENV = prev.nodeEnv
  })

  it('honors explicit true/1 and false/0', () => {
    process.env.NODE_ENV = 'production'
    process.env[KEY] = 'true'
    expect(publicFeatureFlag(KEY)).toBe(true)
    process.env[KEY] = '1'
    expect(publicFeatureFlag(KEY)).toBe(true)
    process.env[KEY] = 'false'
    expect(publicFeatureFlag(KEY)).toBe(false)
    process.env[KEY] = '0'
    expect(publicFeatureFlag(KEY)).toBe(false)
  })

  it('defaults on in non-production when unset', () => {
    delete process.env[KEY]
    process.env.NODE_ENV = 'development'
    expect(publicFeatureFlag(KEY)).toBe(true)
  })

  it('defaults off in production when unset', () => {
    delete process.env[KEY]
    process.env.NODE_ENV = 'production'
    expect(publicFeatureFlag(KEY)).toBe(false)
  })
})

describe('isPublicFlagOn', () => {
  it('treats boolean and string true/1 as on', () => {
    expect(isPublicFlagOn(true)).toBe(true)
    expect(isPublicFlagOn('true')).toBe(true)
    expect(isPublicFlagOn('1')).toBe(true)
    expect(isPublicFlagOn(1)).toBe(true)
  })

  it('rejects falsey and the string "false" (Nuxt env override trap)', () => {
    expect(isPublicFlagOn(false)).toBe(false)
    expect(isPublicFlagOn('false')).toBe(false)
    expect(isPublicFlagOn('0')).toBe(false)
    expect(isPublicFlagOn(0)).toBe(false)
    expect(isPublicFlagOn('')).toBe(false)
    expect(isPublicFlagOn(undefined)).toBe(false)
    expect(isPublicFlagOn(null)).toBe(false)
    expect(Boolean('false')).toBe(true) // document why not Boolean()
  })
})
