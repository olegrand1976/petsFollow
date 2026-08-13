import { describe, expect, it } from 'vitest'
import { applyEidIdentityToForm, type EidFormTarget } from '../../utils/eid-prefill'

describe('applyEidIdentityToForm', () => {
  it('maps identity + address into client create fields', () => {
    const form: EidFormTarget = {
      firstName: '',
      lastName: '',
      address: '',
      nationalRegistryNumber: '',
      billingCustomerKind: '',
      billingCountry: '',
      billingStreet: '',
      billingPostal: '',
      billingCity: '',
    }
    const filled = applyEidIdentityToForm(form, {
      firstname: 'Camille',
      lastname: 'Testeur',
      niss: '96072399828',
      country: 'BE',
      address_street: 'Rue Demo',
      address_number: '1',
      address_zip: '1000',
      address_city: 'Bruxelles',
    })

    expect(form.firstName).toBe('Camille')
    expect(form.lastName).toBe('Testeur')
    expect(form.nationalRegistryNumber).toBe('96072399828')
    expect(form.address).toBe('Rue Demo 1')
    expect(form.billingStreet).toBe('Rue Demo 1')
    expect(form.billingPostal).toBe('1000')
    expect(form.billingCity).toBe('Bruxelles')
    expect(form.billingCountry).toBe('BE')
    expect(form.billingCustomerKind).toBe('individual')
    expect(filled).toContain('firstName')
    expect(filled).toContain('nationalRegistryNumber')
  })

  it('does not overwrite an existing billingCustomerKind', () => {
    const form: EidFormTarget = { billingCustomerKind: 'business' }
    applyEidIdentityToForm(form, { firstname: 'A', lastname: 'B', niss: '96072399828' })
    expect(form.billingCustomerKind).toBe('business')
  })

  it('sets billingCountry BE from niss alone (Web eID without address)', () => {
    const form: EidFormTarget = {}
    applyEidIdentityToForm(form, { firstname: 'A', lastname: 'B', niss: '96072399828' })
    expect(form.billingCountry).toBe('BE')
    expect(form.address).toBeUndefined()
  })

  it('ignores empty fields', () => {
    const form: EidFormTarget = { firstName: 'Keep' }
    applyEidIdentityToForm(form, { firstname: '', lastname: 'Only' })
    expect(form.firstName).toBe('Keep')
    expect(form.lastName).toBe('Only')
  })
})
