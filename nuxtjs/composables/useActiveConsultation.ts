/**
 * Singleton — consultation walk-in active (shell) + reprise après veille/switch.
 * Pas de texte CR en localStorage (PHI) : seulement visitId/clientId/petId/path.
 */
export type ConsultationResume = {
  visitId: string
  clientId: string
  petId: string
  path: string
  updatedAt: number
  /** Consultation ouverte sur un RDV agenda existant — ne pas annuler la visite à la fermeture. */
  keepVisit?: boolean
}

const RESUME_KEY = 'pf_consult_resume'
const TTL_MS = 12 * 60 * 60 * 1000

type ResumeMap = Record<string, ConsultationResume>

/** Module singletons — shared across all useActiveConsultation() callers. */
let flushHandler: (() => Promise<void>) | null = null
let activeSnapshot: { visitId: string, clientId: string, petId: string, keepVisit?: boolean } | null = null

function canUseStorage(): boolean {
  return typeof localStorage !== 'undefined'
}

function readResumeMap(): ResumeMap {
  if (!canUseStorage()) return {}
  try {
    const raw = localStorage.getItem(RESUME_KEY)
    if (!raw) return {}
    return JSON.parse(raw) as ResumeMap
  }
  catch {
    return {}
  }
}

function writeResumeMap(map: ResumeMap) {
  if (!canUseStorage()) return
  try {
    localStorage.setItem(RESUME_KEY, JSON.stringify(map))
  }
  catch {
    /* private mode / quota */
  }
}

function isFresh(token: ConsultationResume): boolean {
  return Date.now() - (token.updatedAt || 0) <= TTL_MS
}

export function useActiveConsultation() {
  const open = useState('pf-consult-open', () => false)
  const clientId = useState('pf-consult-client-id', () => '')
  const resumeVisitId = useState('pf-consult-resume-visit', () => '')
  const resumePetId = useState('pf-consult-resume-pet', () => '')
  /** True quand la consultation porte sur un RDV agenda existant (openForVisit). */
  const resumeKeepVisit = useState('pf-consult-keep-visit', () => false)
  /** Skip discardIfUnsaved on unmount (desk lock / switch). */
  const suspendDiscard = useState('pf-consult-suspend-discard', () => false)

  function registerFlush(fn: (() => Promise<void>) | null) {
    flushHandler = fn
  }

  /** Open walk-in for a client; optional preferredPetId pre-selects the animal (no visit resume). */
  function openForClient(id: string, preferredPetId?: string) {
    if (!id) return
    resumeVisitId.value = ''
    resumePetId.value = preferredPetId || ''
    resumeKeepVisit.value = false
    clientId.value = id
    open.value = true
  }

  /**
   * Open the consultation modal on an existing visit (agenda / historique / reprise).
   * Même coque full que « Nouvelle consultation » après Démarrer.
   */
  function openForVisit(input: { visitId: string, clientId: string, petId?: string, keepVisit?: boolean }) {
    if (!input.visitId || !input.clientId) return
    resumeVisitId.value = input.visitId
    resumePetId.value = input.petId || ''
    resumeKeepVisit.value = input.keepVisit !== false
    clientId.value = input.clientId
    open.value = true
    syncActiveVisit({
      visitId: input.visitId,
      clientId: input.clientId,
      petId: input.petId || '',
      keepVisit: resumeKeepVisit.value,
    })
  }

  /** Reprise desk-lock / veille → même modale CR. */
  function openResume(token: ConsultationResume) {
    if (!token.visitId || !token.clientId) return
    openForVisit({
      visitId: token.visitId,
      clientId: token.clientId,
      petId: token.petId,
      keepVisit: !!token.keepVisit,
    })
  }

  function close() {
    open.value = false
    clientId.value = ''
    resumeVisitId.value = ''
    resumePetId.value = ''
    resumeKeepVisit.value = false
  }

  function saveResumeForEmail(email: string, token: ConsultationResume) {
    if (!email || !token.visitId) return
    const map = readResumeMap()
    map[email.toLowerCase()] = { ...token, updatedAt: Date.now() }
    writeResumeMap(map)
  }

  function clearResumeForEmail(email: string) {
    if (!email) return
    const map = readResumeMap()
    delete map[email.toLowerCase()]
    writeResumeMap(map)
  }

  /** Read without deleting (desk unlock / shell bootstrap). */
  function peekResumeForEmail(email: string): ConsultationResume | null {
    if (!email) return null
    const map = readResumeMap()
    const key = email.toLowerCase()
    const token = map[key]
    if (!token?.visitId || !token.clientId) return null
    if (!isFresh(token)) {
      delete map[key]
      writeResumeMap(map)
      return null
    }
    return token
  }

  /** Read + delete resume token. Prefer peek + clear after successful openResume. */
  function consumeResumeForEmail(email: string): ConsultationResume | null {
    const token = peekResumeForEmail(email)
    if (!token) return null
    clearResumeForEmail(email)
    return token
  }

  function resolveSnap(): { visitId: string, clientId: string, petId: string, keepVisit?: boolean } | null {
    if (activeSnapshot?.visitId && activeSnapshot.clientId) {
      return activeSnapshot
    }
    if (resumeVisitId.value && clientId.value) {
      return {
        visitId: resumeVisitId.value,
        clientId: clientId.value,
        petId: resumePetId.value || '',
        keepVisit: resumeKeepVisit.value,
      }
    }
    return null
  }

  /** Tab hide: flush CR only — do not write resume token (desk lock owns that). */
  async function autosaveOnly(): Promise<boolean> {
    if (!flushHandler) return true
    try {
      await flushHandler()
      return true
    }
    catch {
      return false
    }
  }

  async function runFlushWithRetry(): Promise<boolean> {
    let ok = await autosaveOnly()
    if (!ok) {
      ok = await autosaveOnly()
    }
    return ok
  }

  /**
   * Desk lock/switch: persist then clear UI so another profile cannot see the modal.
   * Always set suspendDiscard when a visit was in progress (even if snapshot race).
   * Returns false if CR flush failed after retry (caller may still lock for security).
   */
  async function flushBeforeSuspend(email: string | undefined, path: string): Promise<boolean> {
    const snapBefore = resolveSnap()
    const hadOpenConsult = Boolean(
      snapBefore?.visitId
      || (open.value && (resumeVisitId.value || clientId.value)),
    )
    const flushOk = await runFlushWithRetry()
    const snap = resolveSnap() || snapBefore
    if (email && snap?.visitId && snap.clientId) {
      saveResumeForEmail(email, {
        visitId: snap.visitId,
        clientId: snap.clientId,
        petId: snap.petId || '',
        path: path || '/',
        updatedAt: Date.now(),
        keepVisit: !!snap.keepVisit,
      })
      suspendDiscard.value = true
    }
    else if (hadOpenConsult) {
      suspendDiscard.value = true
    }
    open.value = false
    clientId.value = ''
    resumeVisitId.value = ''
    resumePetId.value = ''
    resumeKeepVisit.value = false
    activeSnapshot = null
    return flushOk
  }

  function syncActiveVisit(snap: { visitId: string, clientId: string, petId: string, keepVisit?: boolean } | null) {
    activeSnapshot = snap && snap.visitId ? snap : null
  }

  function getActiveVisitSnapshot() {
    return activeSnapshot
  }

  /** After login/shell hydrate: reopen suspended consult for this email once. */
  function tryResumeForCurrentUser(email: string | undefined | null) {
    if (!email || open.value) return false
    const token = peekResumeForEmail(email)
    if (!token) return false
    openResume(token)
    clearResumeForEmail(email)
    return true
  }

  return {
    open,
    clientId,
    resumeVisitId,
    resumePetId,
    resumeKeepVisit,
    suspendDiscard,
    registerFlush,
    syncActiveVisit,
    getActiveVisitSnapshot,
    openForClient,
    openForVisit,
    openResume,
    close,
    saveResumeForEmail,
    clearResumeForEmail,
    peekResumeForEmail,
    consumeResumeForEmail,
    flushBeforeSuspend,
    autosaveOnly,
    tryResumeForCurrentUser,
  }
}
