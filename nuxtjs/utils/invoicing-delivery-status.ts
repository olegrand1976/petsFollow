/**
 * Statut d'acheminement d'un document Billit (`peppolStatus`).
 *
 * La valeur stockée est technique et sert l'audit : le préfixe `email_` marque
 * un envoi SMTP à un particulier, par opposition au réseau Peppol/SDI. Affichée
 * telle quelle, elle donne `email_delivered` ou `stale_timeout` à lire au véto.
 * Extrait de la page pour être testé seul, et pour que l'ajout d'un statut côté
 * Go se voie ici (repli sur la valeur brute plutôt que clé i18n manquante).
 */

/** Préfixe posé par le service Go (`invoicing.EmailStatusPrefix`). */
const EMAIL_PREFIX = 'email_'

/**
 * Statuts du catalogue i18n (`invoicing.deliveryStatus.*`), limités à ceux qui
 * peuvent réellement s'afficher dans la liste d'un cabinet. Deux valeurs écrites
 * par le Go en sont volontairement absentes : `saas_draft` (Flux A, filtré de la
 * liste par `source = 'practice'`) et `accepted` (posé en même temps que le
 * statut de document homonyme, donc toujours tu par `isRedundantDelivery`).
 */
const KNOWN = new Set([
  'creating_order',
  'sending',
  'delivered',
  'rejected',
  'cancelled',
  'unknown',
  'send_failed',
  // Refus de compte Billit (identité non validée) : à distinguer d'un échec
  // technique, le cabinet a une action à faire avant de renvoyer.
  'account_unverified',
  'stale_timeout',
  // Proforma envoyée au client : le document est `issued`, l'attente n'est dite qu'ici.
  'awaiting_client',
])

export type DeliveryStatus = {
  /** Valeur technique d'origine, à garder accessible pour le support. */
  raw: string
  /** Envoi par email (particulier) plutôt que par le réseau e-invoice. */
  byEmail: boolean
  /** Statut sans le canal, à comparer au statut du document. */
  base: string
  /** Clé i18n du statut, `null` si la valeur est inconnue du catalogue. */
  key: string | null
}

/** `null` quand il n'y a rien à afficher (document jamais envoyé). */
export function parseDeliveryStatus(raw: unknown): DeliveryStatus | null {
  const value = String(raw ?? '').trim()
  if (!value) return null
  const byEmail = value.startsWith(EMAIL_PREFIX)
  const base = byEmail ? value.slice(EMAIL_PREFIX.length) : value
  return {
    raw: value,
    byEmail,
    base,
    key: KNOWN.has(base) ? `invoicing.deliveryStatus.${base}` : null,
  }
}

/**
 * Vrai quand la mention ne dirait rien de plus que le statut du document —
 * « Livrée · Remise confirmée » sur une facture Peppol, ou deux fois « Accepté
 * par le client » sur une proforma. Un envoi email n'est jamais redondant : le
 * canal n'apparaît nulle part ailleurs sur la ligne.
 */
export function isRedundantDelivery(status: DeliveryStatus, docStatus: string): boolean {
  return !status.byEmail && status.base === String(docStatus || '').trim()
}
