import { INVOICING_UI_ENABLED } from '~/utils/invoicing-ui'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

/**
 * Orchestration consultation (Nuxt Pro).
 * Setup (modal) → workspace `/consultations/{id}` (CR + hub DAF/facture).
 */
export type ConsultationPet = {
  id: string
  name: string
  species?: string
}

export type StartConsultationInput = {
  clientId: string
  petId: string
  notes?: string
  durationMinutes?: number
}

function isConsultationHasReportError(e: any): boolean {
  const err = e?.data?.error
  const msgKey = err?.msgKey || err?.messageKey
  return msgKey === 'consultation_has_report'
}

export function useConsultationFlow() {
  const config = useRuntimeConfig()
  const pharmacyEnabled = computed(() => isPublicFlagOn(config.public.pharmacyEnabled))
  const billitEnabled = computed(() => isPublicFlagOn(config.public.billitEnabled))
  const invoicingUiEnabled = INVOICING_UI_ENABLED

  const visitId = ref('')
  const petId = ref('')
  const clientId = ref('')
  /** ISO date of the visit — used by ProVisitReportPanel to show/prefix "Date du : …". */
  const scheduledAt = ref('')
  const starting = ref(false)
  const reportSaved = ref(false)
  /** True quand la consultation porte sur un RDV agenda existant — ne jamais l'annuler à la fermeture. */
  const preserveVisit = ref(false)
  /** True while ProVisitReportPanel save/improve/finalize/audio is in flight. */
  const reportBusy = ref(false)
  const error = ref('')

  let reportIdleWaiters: Array<() => void> = []

  function notifyReportIdle() {
    const waiters = reportIdleWaiters
    reportIdleWaiters = []
    for (const resolve of waiters) resolve()
  }

  function setReportBusy(busy: boolean) {
    reportBusy.value = busy
    if (!busy) notifyReportIdle()
  }

  function waitUntilReportIdle(): Promise<void> {
    if (!reportBusy.value) return Promise.resolve()
    return new Promise((resolve) => {
      reportIdleWaiters.push(resolve)
    })
  }

  function reset() {
    // Unblock any discard waiting on reportBusy before clearing state.
    reportBusy.value = false
    notifyReportIdle()
    visitId.value = ''
    petId.value = ''
    clientId.value = ''
    scheduledAt.value = ''
    starting.value = false
    reportSaved.value = false
    preserveVisit.value = false
    error.value = ''
  }

  /** Étape 1 — visite confirmée immédiate (consultation cabinet). */
  async function startConsultation(input: StartConsultationInput) {
    starting.value = true
    error.value = ''
    reportSaved.value = false
    visitId.value = ''
    clientId.value = input.clientId
    petId.value = input.petId
    try {
      const scheduledAtIso = new Date().toISOString()
      scheduledAt.value = scheduledAtIso
      const duration = [15, 30, 60].includes(input.durationMinutes ?? 0)
        ? input.durationMinutes
        : 30
      const res: any = await $fetch(`/api/pets/${input.petId}/visits`, {
        method: 'POST',
        body: {
          confirmDirect: true,
          silentConfirm: true,
          consultationSession: true,
          scheduledAt: scheduledAtIso,
          durationMinutes: duration,
          notes: input.notes?.trim() || undefined,
        },
      })
      const data = res?.data ?? res
      const id = data?.id as string | undefined
      if (!id) {
        throw new Error('missing_visit_id')
      }
      visitId.value = id
      return id
    }
    catch (e: any) {
      error.value = e?.data?.error?.msgKey || e?.data?.error?.code || e?.message || 'error'
      throw e
    }
    finally {
      starting.value = false
    }
  }

  /**
   * Cancel orphan visit if the modal closes before a CR save.
   * Waits for any in-flight CR write so we don't race the PUT;
   * treats server 409 (consultation_has_report) as "already saved".
   */
  async function discardIfUnsaved() {
    // Rien à annuler (CR déjà sauvé / RDV agenda) — ne pas attendre un idle qui peut ne jamais venir.
    if (!visitId.value || reportSaved.value || preserveVisit.value) return
    await waitUntilReportIdle()
    const id = visitId.value
    if (!id || reportSaved.value || preserveVisit.value) return
    try {
      await $fetch(`/api/visits/${id}`, {
        method: 'PATCH',
        body: { status: 'cancelled' },
      })
    }
    catch (e: any) {
      if (isConsultationHasReportError(e)) {
        reportSaved.value = true
        return
      }
      // best-effort cleanup
    }
  }

  async function markDone() {
    const id = visitId.value
    if (!id) return
    await $fetch(`/api/visits/${id}`, {
      method: 'PATCH',
      body: { status: 'done' },
    })
  }

  /** Étape 3 — après save/finalize CR. */
  function afterSaved() {
    reportSaved.value = true
  }

  /** Étape 4a — DAF prérempli (client / pet / visit). */
  function dafWizardPath() {
    const q = new URLSearchParams()
    if (clientId.value) q.set('clientUserId', clientId.value)
    if (petId.value) q.set('petId', petId.value)
    if (visitId.value) q.set('visitId', visitId.value)
    const qs = q.toString()
    return qs ? `/daf/nouveau?${qs}` : '/daf/nouveau'
  }

  /** Étape 4b — facture Billit (direct ou depuis DAF). */
  function invoicingPath(opts?: { dafId?: string, mode?: 'direct' | 'fromDaf' }) {
    const q = new URLSearchParams()
    if (clientId.value) q.set('clientUserId', clientId.value)
    if (visitId.value) q.set('visitId', visitId.value)
    if (opts?.dafId) q.set('dafId', opts.dafId)
    q.set('mode', opts?.mode || (opts?.dafId ? 'fromDaf' : 'direct'))
    return `/invoicing?${q.toString()}`
  }

  /** Étape 4c — nouvelle prescription préremplie (pet / visit). */
  function prescriptionWizardPath() {
    const q = new URLSearchParams()
    if (petId.value) q.set('petId', petId.value)
    if (clientId.value) q.set('clientUserId', clientId.value)
    if (visitId.value) q.set('visitId', visitId.value)
    const qs = q.toString()
    return qs ? `/prescriptions/nouveau?${qs}` : '/prescriptions/nouveau'
  }

  async function loadClientPets(clientUserId: string): Promise<ConsultationPet[]> {
    const res: any = await $fetch(`/api/clients/${clientUserId}/pets`)
    const rows = (res?.data ?? res) as any[]
    if (!Array.isArray(rows)) return []
    return rows.map(p => ({
      id: String(p.id),
      name: String(p.name || ''),
      species: p.species ? String(p.species) : undefined,
    }))
  }

  return {
    pharmacyEnabled,
    billitEnabled,
    invoicingUiEnabled,
    visitId,
    petId,
    clientId,
    scheduledAt,
    starting,
    reportSaved,
    preserveVisit,
    reportBusy,
    error,
    reset,
    startConsultation,
    afterSaved,
    setReportBusy,
    discardIfUnsaved,
    markDone,
    dafWizardPath,
    invoicingPath,
    prescriptionWizardPath,
    loadClientPets,
  }
}
