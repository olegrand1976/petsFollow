/**
 * Calcul des montants lignes facture (centimes) — même arrondi que l’UI
 * `/invoicing` et le body POST. Extrait pour verrouiller HT / TVA / TTC
 * (régression = écart client vs Billit).
 */

export type InvoiceLineInput = {
  description?: string
  quantity: string | number
  unitPriceExcl: string | number
  vatPercent: number | string
}

export type InvoiceLineAmounts = {
  excl: number
  vat: number
  incl: number
}

export type InvoicePayloadLine = {
  description: string
  quantity: number
  unitPriceExclCents: number
  vatPercent: number
}

/** TVA par défaut du pays de facturation (IT = 22, sinon 21). */
export function defaultVatPercent(country: string): number {
  return country === 'IT' ? 22 : 21
}

/** Adresse obligatoire : particulier (mention légale) ou BE pro. */
export function isInvoiceAddressRequired(customerKind: string, country: string): boolean {
  return customerKind === 'individual' || country === 'BE'
}

const VAT_REQUIRED_COUNTRIES = ['BE', 'FR', 'IT'] as const

/** N° TVA obligatoire pour un pro (ES accepte NIF/CIF à la place). */
export function isInvoiceVatRequired(customerKind: string, country: string): boolean {
  return customerKind === 'business' && (VAT_REQUIRED_COUNTRIES as readonly string[]).includes(country)
}

/**
 * Montants d’une ligne en centimes, ou `null` si la ligne ne serait pas
 * envoyée (qty / PU / TVA invalides, ou description vide).
 * Aligné sur `buildInvoicePayloadLines` pour que totaux UI = body POST.
 */
export function lineAmountsCents(line: InvoiceLineInput): InvoiceLineAmounts | null {
  const description = String(line.description ?? '').trim()
  if (!description) return null
  const qty = Number(line.quantity)
  const unitCents = Math.round(Number(line.unitPriceExcl) * 100)
  const vatPercent = Number(line.vatPercent)
  if (!Number.isFinite(qty) || qty <= 0) return null
  if (!Number.isFinite(unitCents) || unitCents <= 0) return null
  if (!Number.isFinite(vatPercent) || vatPercent < 0) return null
  const excl = Math.round(unitCents * qty)
  const vat = Math.round((excl * vatPercent) / 100)
  return { excl, vat, incl: excl + vat }
}

export function sumInvoiceLineTotals(lines: readonly InvoiceLineInput[]): InvoiceLineAmounts {
  let excl = 0
  let vat = 0
  for (const line of lines) {
    const amounts = lineAmountsCents(line)
    if (!amounts) continue
    excl += amounts.excl
    vat += amounts.vat
  }
  return { excl, vat, incl: excl + vat }
}

/** Lignes prêtes pour `POST /invoicing/documents` (filtre les incomplètes). */
export function buildInvoicePayloadLines(lines: readonly InvoiceLineInput[]): InvoicePayloadLine[] {
  return lines
    .map((line) => {
      const amounts = lineAmountsCents(line)
      if (!amounts) return null
      return {
        description: String(line.description ?? '').trim(),
        quantity: Number(line.quantity),
        unitPriceExclCents: Math.round(Number(line.unitPriceExcl) * 100),
        vatPercent: Number(line.vatPercent),
      }
    })
    .filter((l): l is InvoicePayloadLine => l != null)
}

export function formatInvoiceMoney(cents: number): string {
  return `${(cents / 100).toFixed(2)} €`
}
