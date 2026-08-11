import { describe, expect, it } from 'vitest'
import { googleLoginBody } from '../../utils/googleLoginBody'

describe('googleLoginBody', () => {
  it('sends consent:false for returning Pro (existing account path)', () => {
    expect(googleLoginBody('tok-abc', false)).toEqual({
      idToken: 'tok-abc',
      consent: false,
    })
  })

  it('sends consent:true when CGU checked (create-if-absent)', () => {
    expect(googleLoginBody('tok-abc', true)).toEqual({
      idToken: 'tok-abc',
      consent: true,
    })
  })
})
