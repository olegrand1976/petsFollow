import { INVOICING_UI_ENABLED } from '~/utils/invoicing-ui'

/**
 * Orchestration « Nouvelle Consultation » (Nuxt Pro).
 * Étapes : 1) créer visite confirmée → 2) CR (ProVisitReportPanel) →
 * 3) CTA post-save → 4) deep-link DAF et/ou facture Billit.
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
  const status = e?.statusCode ?? e?.status ?? e?.response?.status
  if (status === 409) return true
  const key = e?.data?.error?.msgKey || e?.data?.error?.code || e?.data?.error?.messageKey
  return key === 'consultation_has_report'
}

export function useConsultationFlow() {
  const config = useRuntimeConfig()
  const pharmacyEnabled = computed(() => Boolean(config.public.pharmacyEnabled))
  const billitEnabled = computed(() => Boolean(config.public.billitEnabled))
  const invoicingUiEnabled = INVOICING_UI_ENABLED

  const visitId = ref('')
  const petId = ref('')
  const clientId = ref('')
  const starting = ref(false)
  const reportSaved = ref(false)
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
    visitId.value = ''
    petId.value = ''
    clientId.value = ''
    starting.value = false
    reportSaved.value = false
    reportBusy.value = false
    reportIdleWaiters = []
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
      const scheduledAt = new Date().toISOString()
      const duration = [15, 30, 60].includes(input.durationMinutes ?? 0)
        ? input.durationMinutes
        : 30
      const res: any = await $fetch(`/api/pets/${input.petId}/visits`, {
        method: 'POST',
        body: {
          confirmDirect: true,
          silentConfirm: true,
          consultationSession: true,
          scheduledAt,
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
    await waitUntilReportIdle()
    const id = visitId.value
    if (!id || reportSaved.value) return
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
    starting,
    reportSaved,
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
    loadClientPets,
  }
}
