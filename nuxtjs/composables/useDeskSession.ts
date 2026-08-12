import {
  clearAuthTokens,
  finishClientLoginSession,
  hasSessionCookie,
  homePathForRole,
  isAuthSuccess,
  isMFAChallenge,
  isPracticeStaffRole,
  isProRole,
  markAuthSessionActive,
  unwrapAuthData,
} from '~/composables/useAuth'
import { useActiveConsultation } from '~/composables/useActiveConsultation'
import {
  DEFAULT_DESK_IDLE_MINUTES,
  normalizeDeskIdleMinutes,
  resolveDeskIdleMinutes,
} from '~/utils/deskIdleMinutes'

export type DeskMember = {
  email: string
  fullName: string
  teamRole: string
}

export type DeskPromptMode = 'lock' | 'switch' | null

export type DeskAuthResult =
  | { ok: true; role: string | null }
  | { ok: false; reason: 'mfa'; mfaToken: string }
  | { ok: false; reason: 'error' | 'proOnly' }

const ROSTER_KEY = 'pf_desk_roster'
const PATHS_KEY = 'pf_desk_last_paths'
const LOCKED_KEY = 'pf_desk_locked'
const IDLE_MINUTES_KEY = 'pf_desk_idle_minutes'
const DEFAULT_IDLE_MS = DEFAULT_DESK_IDLE_MINUTES * 60 * 1000

/** Module-singleton idle watch — shared across all useDeskSession() callers. */
let idleTimer: ReturnType<typeof setTimeout> | null = null
let idleStarted = false

type RosterCache = {
  practiceId: string
  members: DeskMember[]
  deskIdleMinutes?: number
  updatedAt: number
}

function readJson<T>(key: string, fallback: T): T {
  if (!import.meta.client) return fallback
  try {
    const raw = localStorage.getItem(key)
    if (!raw) return fallback
    return JSON.parse(raw) as T
  } catch {
    return fallback
  }
}

function writeJson(key: string, value: unknown) {
  if (!import.meta.client) return
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch {
    /* private mode / quota */
  }
}

function idleMinutesStorageKey(practiceId?: string): string {
  const pid = (practiceId || '').trim()
  return pid ? `${IDLE_MINUTES_KEY}:${pid}` : IDLE_MINUTES_KEY
}

function persistDeskIdleMinutes(minutes: number, practiceId?: string) {
  if (!import.meta.client) return
  try {
    const pid = practiceId || readJson<RosterCache | null>(ROSTER_KEY, null)?.practiceId
    localStorage.setItem(idleMinutesStorageKey(pid), String(normalizeDeskIdleMinutes(minutes)))
  } catch {
    /* ignore */
  }
}

function cachedDeskIdleMinutes(practiceId?: string): number {
  if (!import.meta.client) return DEFAULT_DESK_IDLE_MINUTES
  const cache = readJson<RosterCache | null>(ROSTER_KEY, null)
  const pid = practiceId || cache?.practiceId
  let scoped: string | null = null
  let legacy: string | null = null
  try {
    scoped = localStorage.getItem(idleMinutesStorageKey(pid))
    if (pid) {
      legacy = localStorage.getItem(IDLE_MINUTES_KEY)
    }
  } catch {
    /* ignore */
  }
  return resolveDeskIdleMinutes({
    practiceId: pid,
    roster: cache,
    scopedIdleRaw: scoped,
    legacyIdleRaw: legacy,
  })
}

/** Client-only flag used by auth middleware to avoid bounce to /login while locked. */
export function isDeskLockedFlag(): boolean {
  if (!import.meta.client) return false
  try {
    return sessionStorage.getItem(LOCKED_KEY) === '1'
  } catch {
    return false
  }
}

function setDeskLockedFlag(on: boolean) {
  if (!import.meta.client) return
  try {
    if (on) sessionStorage.setItem(LOCKED_KEY, '1')
    else sessionStorage.removeItem(LOCKED_KEY)
  } catch {
    /* ignore */
  }
}

function idleMs(): number {
  if (!import.meta.client) return DEFAULT_IDLE_MS
  // Override only hors prod — évite de désactiver la veille partagée en production.
  const env = String(useRuntimeConfig().public.appEnv || '')
  const allowOverride = import.meta.dev || env === 'staging' || env === 'local' || env === 'test'
  if (allowOverride) {
    const w = window as Window & { __PF_DESK_IDLE_MS?: number }
    if (typeof w.__PF_DESK_IDLE_MS === 'number' && w.__PF_DESK_IDLE_MS > 0) {
      return w.__PF_DESK_IDLE_MS
    }
  }
  return cachedDeskIdleMinutes() * 60 * 1000
}

const activityEvents = ['mousemove', 'mousedown', 'keydown', 'scroll', 'touchstart', 'click'] as const

export function useDeskSession() {
  const { user } = useProUser()
  const route = useRoute()

  const roster = useState<DeskMember[]>('pf-desk-roster', () => [])
  const locked = useState<boolean>('pf-desk-locked', () => false)
  const promptMode = useState<DeskPromptMode>('pf-desk-prompt-mode', () => null)
  const pendingEmail = useState<string>('pf-desk-pending-email', () => '')
  /** True when overlay covers the app (lock or switch) — stop polling / hide PHI. */
  const uiBlocked = computed(() => locked.value || promptMode.value === 'switch' || promptMode.value === 'lock')
  const enabled = computed(() => isPracticeStaffRole(user.value?.role) || locked.value)
  /** Poste partagé : veille idle + switcher seulement si l’équipe a ≥ 2 comptes. */
  const sharedDesk = computed(() => roster.value.length > 1)

  function loadRosterFromCache(practiceId?: string) {
    const cache = readJson<RosterCache | null>(ROSTER_KEY, null)
    if (!cache?.members?.length) return
    if (practiceId && cache.practiceId && cache.practiceId !== practiceId) return
    roster.value = cache.members
  }

  function persistRoster(practiceId: string, members: DeskMember[], deskIdleMinutes?: number) {
    roster.value = members
    const minutes = deskIdleMinutes != null
      ? normalizeDeskIdleMinutes(deskIdleMinutes)
      : cachedDeskIdleMinutes(practiceId)
    persistDeskIdleMinutes(minutes, practiceId)
    writeJson(ROSTER_KEY, {
      practiceId,
      members,
      deskIdleMinutes: minutes,
      updatedAt: Date.now(),
    } satisfies RosterCache)
    syncIdleWatch()
  }

  function lastPaths(): Record<string, string> {
    return readJson<Record<string, string>>(PATHS_KEY, {})
  }

  function saveLastPath(email: string, path: string) {
    if (!email || !path.startsWith('/') || path.startsWith('//')) return
    const map = lastPaths()
    map[email.toLowerCase()] = path
    writeJson(PATHS_KEY, map)
  }

  function getLastPath(email: string): string | null {
    const path = lastPaths()[email.toLowerCase()]
    if (!path || !path.startsWith('/') || path.startsWith('//')) return null
    return path
  }

  function rememberCurrentPath() {
    const email = user.value?.email
    if (!email || locked.value) return
    const full = route.fullPath || route.path
    if (full.startsWith('/login')) return
    saveLastPath(email, full)
  }

  function unwrapTeamPayload(res: unknown): { members: DeskMember[]; deskIdleMinutes: number } {
    const root = res as { data?: unknown } | unknown
    const data = root && typeof root === 'object' && 'data' in root ? (root as { data: unknown }).data : root
    if (Array.isArray(data)) {
      return {
        members: data
          .filter((m: { email?: string }) => !!m.email)
          .map((m: { email?: string; fullName?: string; teamRole?: string }) => ({
            email: String(m.email),
            fullName: String(m.fullName || m.email),
            teamRole: String(m.teamRole || ''),
          })),
        deskIdleMinutes: cachedDeskIdleMinutes(),
      }
    }
    const obj = (data && typeof data === 'object' ? data : {}) as {
      members?: Array<{ email?: string; fullName?: string; teamRole?: string }>
      deskIdleMinutes?: number
    }
    const list = Array.isArray(obj.members) ? obj.members : []
    return {
      members: list
        .filter((m) => !!m.email)
        .map((m) => ({
          email: String(m.email),
          fullName: String(m.fullName || m.email),
          teamRole: String(m.teamRole || ''),
        })),
      deskIdleMinutes: normalizeDeskIdleMinutes(obj.deskIdleMinutes),
    }
  }

  async function refreshRoster() {
    if (!isPracticeStaffRole(user.value?.role) || !user.value?.practiceId) return
    try {
      const res = await $fetch('/api/vet/team')
      const { members, deskIdleMinutes } = unwrapTeamPayload(res)
      persistRoster(user.value.practiceId, members, deskIdleMinutes)
    } catch {
      loadRosterFromCache(user.value.practiceId)
      syncIdleWatch()
    }
  }

  function clearIdleTimer() {
    if (idleTimer) {
      clearTimeout(idleTimer)
      idleTimer = null
    }
  }

  function bumpIdle() {
    if (!import.meta.client || locked.value || promptMode.value || !isPracticeStaffRole(user.value?.role)) return
    // Solo cabinet : pas de veille applicative (poste non partagé).
    if (!sharedDesk.value) {
      clearIdleTimer()
      return
    }
    clearIdleTimer()
    idleTimer = setTimeout(() => {
      void lock()
    }, idleMs())
  }

  function onActivity() {
    bumpIdle()
  }

  function startIdleWatch() {
    if (!import.meta.client || !sharedDesk.value) {
      clearIdleTimer()
      return
    }
    if (idleStarted) {
      bumpIdle()
      return
    }
    idleStarted = true
    for (const ev of activityEvents) {
      window.addEventListener(ev, onActivity, { passive: true })
    }
    bumpIdle()
  }

  function stopIdleWatch() {
    if (!import.meta.client || !idleStarted) {
      clearIdleTimer()
      return
    }
    idleStarted = false
    for (const ev of activityEvents) {
      window.removeEventListener(ev, onActivity)
    }
    clearIdleTimer()
  }

  /** Démarre ou coupe le timer selon roster (≥2) + rôle + pas déjà en veille. */
  function syncIdleWatch() {
    if (!import.meta.client || locked.value || promptMode.value) {
      clearIdleTimer()
      return
    }
    if (!isPracticeStaffRole(user.value?.role) || !sharedDesk.value) {
      stopIdleWatch()
      return
    }
    startIdleWatch()
  }

  async function lock() {
    if (locked.value) return
    const email = user.value?.email
    const full = route.fullPath || route.path
    const { flushBeforeSuspend } = useActiveConsultation()
    // Security first: always lock even if CR flush fails (resume token still written when possible).
    await flushBeforeSuspend(email, full)
    rememberCurrentPath()
    clearIdleTimer()
    locked.value = true
    setDeskLockedFlag(true)
    promptMode.value = 'lock'
    pendingEmail.value = email || roster.value[0]?.email || ''
    loadRosterFromCache(user.value?.practiceId)
    await clearAuthTokens()
  }

  function emailAllowedForDesk(email: string): boolean {
    if (!roster.value.length) return true
    const needle = email.toLowerCase()
    return roster.value.some((m) => m.email.toLowerCase() === needle)
  }

  /** Switch = suspend session (comme veille) : autre onglet ne doit plus voir le JWT précédent. */
  async function openSwitch(email: string) {
    if (!email || email.toLowerCase() === user.value?.email?.toLowerCase()) return
    if (!sharedDesk.value) return
    const currentEmail = user.value?.email
    const full = route.fullPath || route.path
    const { flushBeforeSuspend } = useActiveConsultation()
    await flushBeforeSuspend(currentEmail, full)
    rememberCurrentPath()
    clearIdleTimer()
    const practiceId = user.value?.practiceId
    pendingEmail.value = email
    // locked AVANT purge — Annuler ne doit jamais restaurer sans re-auth (race logout).
    locked.value = true
    setDeskLockedFlag(true)
    promptMode.value = 'switch'
    loadRosterFromCache(practiceId)
    await clearAuthTokens()
  }

  function cancelPrompt() {
    // Annuler switch ≠ restore : veille + re-auth obligatoire (même si purge encore en cours).
    locked.value = true
    setDeskLockedFlag(true)
    promptMode.value = 'lock'
  }

  /** Exit veille without unlocking a profile — full login page. */
  async function abandonToLogin() {
    locked.value = false
    setDeskLockedFlag(false)
    promptMode.value = null
    pendingEmail.value = ''
    stopIdleWatch()
    await clearAuthTokens()
    await navigateTo('/login')
  }

  async function authenticate(email: string, password: string): Promise<DeskAuthResult> {
    if (!emailAllowedForDesk(email)) return { ok: false, reason: 'error' }
    try {
      const res = await $fetch('/api/auth/login', {
        method: 'POST',
        body: { email, password },
      })
      const data = unwrapAuthData(res)
      if (isMFAChallenge(data)) {
        return { ok: false, reason: 'mfa', mfaToken: data.mfaToken }
      }
      if (!isAuthSuccess(data)) return { ok: false, reason: 'error' }
      const role = typeof data.role === 'string' ? data.role : null
      if (role && !isProRole(role)) {
        await clearAuthTokens()
        return { ok: false, reason: 'proOnly' }
      }
      return { ok: true, role }
    } catch {
      return { ok: false, reason: 'error' }
    }
  }

  async function verify2fa(mfaToken: string, code: string): Promise<DeskAuthResult> {
    try {
      const res = await $fetch('/api/auth/2fa/verify', {
        method: 'POST',
        body: { mfaToken, code },
      })
      const data = unwrapAuthData(res)
      if (isMFAChallenge(data)) {
        return { ok: false, reason: 'mfa', mfaToken: data.mfaToken }
      }
      if (!isAuthSuccess(data)) return { ok: false, reason: 'error' }
      const role = typeof data.role === 'string' ? data.role : null
      if (role && !isProRole(role)) {
        await clearAuthTokens()
        return { ok: false, reason: 'proOnly' }
      }
      return { ok: true, role }
    } catch {
      return { ok: false, reason: 'error' }
    }
  }

  function completeUnlock(email: string, role: string | null) {
    const { peekResumeForEmail } = useActiveConsultation()
    const resume = peekResumeForEmail(email)
    const target = resume?.path || getLastPath(email) || (role ? homePathForRole(role) : '/dashboard')
    locked.value = false
    setDeskLockedFlag(false)
    promptMode.value = null
    pendingEmail.value = ''
    markAuthSessionActive()
    // finishClientLoginSession → location.replace (hard reload). Leave the resume
    // token in localStorage so ProActiveConsultationHost.tryResumeForCurrentUser
    // can reopen after shell boot. openResume+clearResume here would wipe useState
    // on reload and drop the token (consultation never resumes).
    finishClientLoginSession(target)
  }

  async function bootstrap() {
    if (!import.meta.client) return
    // Login frais via /login (layout:false) peut laisser pf_desk_locked stale :
    // session httpOnly présente ⇒ ce n'est plus une veille, ne pas re-demander le MDP.
    // !locked : évite de lever une veille volontaire pendant await clearAuthTokens()
    // (flag déjà posé, cookies pas encore purgés — remount layout / double mount).
    if (isDeskLockedFlag() && hasSessionCookie() && !locked.value) {
      setDeskLockedFlag(false)
      locked.value = false
      if (promptMode.value === 'lock' || promptMode.value === 'switch') {
        promptMode.value = null
      }
      pendingEmail.value = ''
    } else if (isDeskLockedFlag()) {
      locked.value = true
      promptMode.value = 'lock'
      loadRosterFromCache()
      stopIdleWatch()
      return
    }
    if (!isPracticeStaffRole(user.value?.role)) {
      stopIdleWatch()
      return
    }
    loadRosterFromCache(user.value.practiceId)
    await refreshRoster()
    syncIdleWatch()
    rememberCurrentPath()
  }

  /** Après sauvegarde settings : propage la durée cabinet au timer idle. */
  function setDeskIdleMinutes(minutes: number) {
    const practiceId = user.value?.practiceId
      || readJson<RosterCache | null>(ROSTER_KEY, null)?.practiceId
    persistDeskIdleMinutes(minutes, practiceId)
    const cache = readJson<RosterCache | null>(ROSTER_KEY, null)
    if (cache) {
      writeJson(ROSTER_KEY, {
        ...cache,
        deskIdleMinutes: normalizeDeskIdleMinutes(minutes),
        updatedAt: Date.now(),
      } satisfies RosterCache)
    }
    syncIdleWatch()
  }

  function forceLockForTests() {
    void lock()
  }

  /** E2E : forcer un roster (ex. solo) et resync le timer idle. */
  function setRosterForTests(members: DeskMember[]) {
    roster.value = Array.isArray(members) ? members : []
    syncIdleWatch()
  }

  if (import.meta.client) {
    const w = window as Window & {
      __PF_DESK_FORCE_LOCK?: () => void
      __PF_DESK_SET_ROSTER?: (members: DeskMember[]) => void
    }
    w.__PF_DESK_FORCE_LOCK = forceLockForTests
    w.__PF_DESK_SET_ROSTER = setRosterForTests
  }

  return {
    roster,
    locked,
    promptMode,
    pendingEmail,
    uiBlocked,
    enabled,
    sharedDesk,
    refreshRoster,
    rememberCurrentPath,
    openSwitch,
    cancelPrompt,
    abandonToLogin,
    lock,
    authenticate,
    verify2fa,
    completeUnlock,
    bootstrap,
    startIdleWatch,
    stopIdleWatch,
    bumpIdle,
    syncIdleWatch,
    setDeskIdleMinutes,
    forceLockForTests,
    setRosterForTests,
  }
}
