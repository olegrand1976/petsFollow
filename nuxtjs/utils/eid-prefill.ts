export type EidIdentity = {
  lastname?: string
  firstname?: string
  firstnames?: string
  birth_date?: string
  birth_place?: string
  gender?: string
  country?: string
  niss?: string
  card_number?: string
  validity_from?: string
  validity_to?: string
  address_street?: string
  address_number?: string
  address_zip?: string
  address_city?: string
  signature_verified?: boolean
  import_tool?: string
  photo_jpeg_base64?: string
}

export type EidFormTarget = {
  firstName?: string
  lastName?: string
  address?: string
  nationalRegistryNumber?: string
  billingCustomerKind?: string
  billingCountry?: string
  billingStreet?: string
  billingPostal?: string
  billingCity?: string
}

/** Prefill client form fields from a normalized eID identity payload. */
export function applyEidIdentityToForm<T extends EidFormTarget>(form: T, data: EidIdentity): (keyof T)[] {
  const filled: (keyof T)[] = []
  const set = <K extends keyof T>(key: K, value: string | undefined) => {
    if (value == null || String(value).trim() === '') return
    form[key] = String(value).trim() as T[K]
    filled.push(key)
  }

  set('firstName' as keyof T, data.firstname)
  set('lastName' as keyof T, data.lastname)
  set('nationalRegistryNumber' as keyof T, data.niss)

  const streetParts = [data.address_street, data.address_number].filter(Boolean)
  const streetLine = streetParts.join(' ').trim()
  if (streetLine) {
    set('address' as keyof T, streetLine)
    set('billingStreet' as keyof T, streetLine)
  }
  set('billingPostal' as keyof T, data.address_zip)
  set('billingCity' as keyof T, data.address_city)

  if (data.country || data.niss) {
    set('billingCountry' as keyof T, (data.country || 'BE').slice(0, 2).toUpperCase())
  }
  const kind = (form as EidFormTarget).billingCustomerKind
  if (!kind) {
    set('billingCustomerKind' as keyof T, 'individual')
  }
  return filled
}
