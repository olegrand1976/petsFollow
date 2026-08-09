/**
 * Préremplissage des lignes de facture (BIL-9) depuis la consultation : acte au
 * tarif du type de RDV + médicaments du DAF finalisé. L'API répond en centimes
 * HTVA, le formulaire saisit des euros. Extrait de `/invoicing` pour être testé
 * seul — une erreur de conversion ici part telle quelle chez le client.
 */

/** Ligne renvoyée par `GET /practices/me/invoicing/prefill`. */
export type PrefillApiLine = {
  description?: string
  quantity?: number
  unitPriceExclCents?: number
  /** Absent (ou `null`) quand le prix est inconnu du catalogue. */
  vatPercent?: number | null
  source?: string
}

export type PrefillPayload = {
  lines?: PrefillApiLine[]
  visitTypeName?: string
  dafId?: string
  clientName?: string
}

/** Ligne prête pour le formulaire (champs texte, comme la saisie manuelle). */
export type PrefillDraftLine = {
  description: string
  quantity: string
  unitPriceExcl: string
  vatPercent: number
}

export type PrefillDraft = {
  lines: PrefillDraftLine[]
  /** Total HTVA des lignes tarifées, en centimes. 0 si rien n'est chiffrable. */
  estimatedExclCents: number
  /** Nombre de lignes venant du DAF, pour l'indice affiché au véto. */
  dafLineCount: number
}

/**
 * `vatDefault` : taux du pays de facturation, appliqué aux lignes dont l'API ne
 * connaît pas la TVA — imposer un 21 % belge fausserait une facture italienne.
 */
export function prefillDraft(payload: PrefillPayload | null | undefined, vatDefault: number): PrefillDraft {
  const raw = payload?.lines
  const lines: PrefillApiLine[] = Array.isArray(raw) ? raw : []
  const out: PrefillDraftLine[] = []
  let estimatedExclCents = 0

  for (const l of lines) {
    const qty = Number(l?.quantity)
    const quantity = Number.isFinite(qty) && qty > 0 ? String(qty) : '1'
    const line: PrefillDraftLine = {
      description: String(l?.description || ''),
      quantity,
      unitPriceExcl: '',
      vatPercent: vatDefault,
    }
    const unitCents = Number(l?.unitPriceExclCents)
    // Prix inconnu du catalogue : champ laissé vide, le véto complète.
    if (Number.isFinite(unitCents) && unitCents > 0) {
      line.unitPriceExcl = (unitCents / 100).toFixed(2)
      estimatedExclCents += Math.round(unitCents * Number(quantity))
    }
    const vat = Number(l?.vatPercent)
    if (l?.vatPercent != null && Number.isFinite(vat) && vat >= 0) line.vatPercent = vat
    out.push(line)
  }

  return {
    lines: out,
    estimatedExclCents,
    dafLineCount: lines.filter((l) => l?.source === 'daf').length,
  }
}
