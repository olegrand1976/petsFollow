/**
 * Tarif d'un type de rendez-vous (BIL-9) : le véto saisit des euros, l'API
 * stocke des centimes HTVA. Extrait de `/settings` pour être testé seul —
 * une erreur d'arrondi ici se retrouve directement sur une facture Billit.
 */

/** Euros saisis → centimes HTVA. `null` = saisie invalide, vide = non tarifé (0). */
export function visitTypePriceCents(priceExcl: unknown): number | null {
  const raw = String(priceExcl ?? '').trim().replace(',', '.')
  if (raw === '') return 0
  const euros = Number(raw)
  if (!Number.isFinite(euros) || euros < 0) return null
  return Math.round(euros * 100)
}

/** Centimes → champ de saisie. 0 s'affiche vide : « non tarifé », pas « 0,00 ». */
export function visitTypePriceInput(cents: unknown): string {
  const n = Number(cents || 0)
  return Number.isFinite(n) && n > 0 ? (n / 100).toFixed(2) : ''
}

/** Mêmes bornes que la validation des lignes de facture côté API. */
export function isValidVisitTypeVat(vatPercent: unknown): boolean {
  const vat = Number(vatPercent)
  return Number.isFinite(vat) && vat >= 0 && vat <= 100
}
