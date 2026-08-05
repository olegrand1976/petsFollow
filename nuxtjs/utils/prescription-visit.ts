/** Helpers to link a pet visit (consultation) to consignes and label it in selects. */
export type PrescriptionVisitOption = {
  id: string
  scheduledAt?: string
  status?: string
  notes?: string
  reportStatus?: string
  hasFinalReport?: boolean
  /** Present in draft/URL but missing from GET visits (soft-deleted / wrong pet). */
  orphan?: boolean
}

export function unwrapApiData(res: any) {
  return res?.data ?? res
}

export function formatPrescriptionVisitLabel(
  v: PrescriptionVisitOption,
  unavailableLabel?: string,
): string {
  if (!v?.id) return ''
  if (v.orphan) return unavailableLabel || '—'
  const when = v.scheduledAt ? String(v.scheduledAt).slice(0, 16).replace('T', ' ') : '—'
  const status = v.status || ''
  const report = v.hasFinalReport || v.reportStatus === 'final'
    ? ' · CR'
    : (v.reportStatus === 'draft' ? ' · CR draft' : '')
  const note = (v.notes || '').trim()
  const noteBit = note ? ` — ${note.slice(0, 40)}${note.length > 40 ? '…' : ''}` : ''
  return `${when}${status ? ` · ${status}` : ''}${report}${noteBit}`
}

/** Append a synthetic option when linkedId is set but absent from the fetched list. */
export function withEnsuredLinkedVisit(
  visits: PrescriptionVisitOption[],
  linkedId: string,
): PrescriptionVisitOption[] {
  const id = String(linkedId || '').trim()
  if (!id) return visits
  if (visits.some(v => v.id === id)) return visits
  return [...visits, { id, orphan: true }]
}

export function findVisitOption(
  visits: PrescriptionVisitOption[],
  linkedId: string,
): PrescriptionVisitOption | undefined {
  const id = String(linkedId || '').trim()
  if (!id) return undefined
  return visits.find(v => v.id === id)
}

/** True when the linked id is missing from the list or marked orphan. */
export function isOrphanLinkedVisit(
  visits: PrescriptionVisitOption[],
  linkedId: string,
): boolean {
  const id = String(linkedId || '').trim()
  if (!id) return false
  const v = findVisitOption(visits, id)
  return !v || !!v.orphan
}

/**
 * visitId to send on create/PATCH: never send an orphan (would 400 visit_mismatch).
 * Empty string clears an existing link on PATCH.
 */
export function visitIdForSave(
  visits: PrescriptionVisitOption[],
  linkedId: string,
): string {
  const id = String(linkedId || '').trim()
  if (!id || isOrphanLinkedVisit(visits, id)) return ''
  return id
}

export async function fetchPetVisitsForPrescription(petId: string): Promise<PrescriptionVisitOption[]> {
  const id = String(petId || '').trim()
  if (!id) return []
  const res = await $fetch<any>(`/api/pets/${id}/visits`)
  const data = unwrapApiData(res)
  const rows = Array.isArray(data) ? data : (data?.items ?? data?.visits ?? [])
  return rows
    .map((v: any) => ({
      id: String(v.id || ''),
      scheduledAt: v.scheduledAt ? String(v.scheduledAt) : undefined,
      status: v.status ? String(v.status) : undefined,
      notes: v.notes ? String(v.notes) : undefined,
      reportStatus: v.reportStatus ? String(v.reportStatus) : undefined,
      hasFinalReport: !!v.hasFinalReport,
    }))
    .filter((v: PrescriptionVisitOption) => !!v.id)
}
