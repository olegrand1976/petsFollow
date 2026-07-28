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
}

const RESUME_KEY = 'pf_consult_resume'
const TTL_MS = 12 * 60 * 60 * 1000

type ResumeMap = Record<string, ConsultationResume>

/** Module singletons — shared across all useActiveConsultation() callers. */
let flushHandler: (() => Promise<void>) | null = null
let activeSnapshot: { visitId: string, clientId: string, petId: string } | null = null

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

export function useActiveConsultation() {
  const open = useState('pf-consult-open', () => false)
  const clientId = useState('pf-consult-client-id', () => '')
  const resumeVisitId = useState('pf-consult-resume-visit', () => '')
  const resumePetId = useState('pf-consult-resume-pet', () => '')
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
    clientId.value = id
    open.value = true
  }

  function openResume(token: ConsultationResume) {
    resumeVisitId.value = token.visitId
    resumePetId.value = token.petId
    clientId.value = token.clientId
    open.value = true
  }

  function close() {
    open.value = false
    clientId.value = ''
    resumeVisitId.value = ''
    resumePetId.value = ''
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

  /** Read + delete resume token (true consume). */
  function consumeResumeForEmail(email: string): ConsultationResume | null {
    if (!email) return null
    const map = readResumeMap()
    const key = email.toLowerCase()
    const token = map[key]
    if (!token?.visitId || !token.clientId) return null
    delete map[key]
    writeResumeMap(map)
    if (Date.now() - (token.updatedAt || 0) > TTL_MS) {
      return null
    }
    return token
  }

  function resolveSnap(): { visitId: string, clientId: string, petId: string } | null {
    if (activeSnapshot?.visitId && activeSnapshot.clientId) {
      return activeSnapshot
    }
    if (resumeVisitId.value && clientId.value) {
      return {
        visitId: resumeVisitId.value,
        clientId: clientId.value,
        petId: resumePetId.value || '',
      }
    }
    return null
  }

  async function autosaveAndRemember(email: string | undefined, path: string): Promise<boolean> {
    let flushOk = true
    if (flushHandler) {
      try {
        await flushHandler()
      }
      catch {
        flushOk = false
      }
    }
    const snap = resolveSnap()
    if (email && snap?.visitId && snap.clientId) {
      saveResumeForEmail(email, {
        visitId: snap.visitId,
        clientId: snap.clientId,
        petId: snap.petId || '',
        path: path || '/',
        updatedAt: Date.now(),
      })
    }
    return flushOk
  }

  /**
   * Desk lock/switch: persist then clear UI so another profile cannot see the modal.
   * Always set suspendDiscard when a visit was in progress (even if snapshot race).
   */
  async function flushBeforeSuspend(email: string | undefined, path: string) {
    const hadOpenConsult = Boolean(open.value && (resolveSnap()?.visitId || resumeVisitId.value))
    await autosaveAndRemember(email, path)
    const snap = resolveSnap()
    if (email && snap?.visitId && snap.clientId) {
      // Re-save after flush (snapshot may have been updated by forceSave path).
      saveResumeForEmail(email, {
        visitId: snap.visitId,
        clientId: snap.clientId,
        petId: snap.petId || '',
        path: path || '/',
        updatedAt: Date.now(),
      })
      suspendDiscard.value = true
    }
    else if (hadOpenConsult) {
      // Snapshot race: still skip orphan cancel on unmount.
      suspendDiscard.value = true
    }
    open.value = false
    clientId.value = ''
    resumeVisitId.value = ''
    resumePetId.value = ''
    activeSnapshot = null
  }

  function syncActiveVisit(snap: { visitId: string, clientId: string, petId: string } | null) {
    activeSnapshot = snap && snap.visitId ? snap : null
  }

  function getActiveVisitSnapshot() {
    return activeSnapshot
  }

  return {
    open,
    clientId,
    resumeVisitId,
    resumePetId,
    suspendDiscard,
    registerFlush,
    syncActiveVisit,
    getActiveVisitSnapshot,
    openForClient,
    openResume,
    close,
    saveResumeForEmail,
    clearResumeForEmail,
    consumeResumeForEmail,
    flushBeforeSuspend,
    autosaveAndRemember,
  }
}
