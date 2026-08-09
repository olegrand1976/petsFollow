export type CustomerKind = 'individual' | 'business'

export type CounterpartyLike = {
  customerKind?: string
  vatNumber?: string
  companyNumber?: string
  siret?: string
  siren?: string
  codiceDestinatario?: string
  pec?: string
  taxId?: string
}

/**
 * Type de client d’une contrepartie réutilisée (note de crédit, fiche client).
 *
 * `customerKind` n’existe que depuis la bascule Particulier / Professionnel :
 * une contrepartie plus ancienne n’en a pas, et la déduire « particulier » ferait
 * sauter la TVA d’un avoir B2B et l’enverrait par email au lieu de Peppol.
 * En son absence, la présence d’un identifiant fiscal fait foi.
 */
export function counterpartyKind(cp: CounterpartyLike): CustomerKind {
  if (cp.customerKind === 'business' || cp.customerKind === 'individual') return cp.customerKind
  const hasFiscalId = Boolean(
    cp.vatNumber || cp.companyNumber || cp.siret || cp.siren || cp.codiceDestinatario || cp.pec || cp.taxId,
  )
  return hasFiscalId ? 'business' : 'individual'
}
