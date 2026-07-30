/** Allowed desk idle durations (minutes) — must stay aligned with Go `kernel.NormalizeDeskIdleMinutes`. */
export const ALLOWED_DESK_IDLE_MINUTES = [1, 2, 5, 10, 15, 30] as const
export const DEFAULT_DESK_IDLE_MINUTES = 2

export function normalizeDeskIdleMinutes(raw: unknown): number {
  const n = Number(raw)
  if ((ALLOWED_DESK_IDLE_MINUTES as readonly number[]).includes(n)) return n
  return DEFAULT_DESK_IDLE_MINUTES
}

export type DeskIdleRosterHint = {
  practiceId?: string
  deskIdleMinutes?: number
}

/**
 * Resolve idle minutes for a shared desk.
 * Server-synced roster value always wins over localStorage (shared-desk hardening).
 * localStorage is only a first-paint / offline fallback when roster has no value yet.
 */
export function resolveDeskIdleMinutes(input: {
  practiceId?: string
  roster?: DeskIdleRosterHint | null
  scopedIdleRaw?: string | null
  legacyIdleRaw?: string | null
}): number {
  const roster = input.roster
  const pid = (input.practiceId || roster?.practiceId || '').trim()

  if (
    roster?.deskIdleMinutes != null
    && (!pid || !roster.practiceId || roster.practiceId === pid)
  ) {
    return normalizeDeskIdleMinutes(roster.deskIdleMinutes)
  }

  if (input.scopedIdleRaw != null) {
    return normalizeDeskIdleMinutes(input.scopedIdleRaw)
  }
  if (
    input.legacyIdleRaw != null
    && (!roster?.practiceId || !pid || roster.practiceId === pid)
  ) {
    return normalizeDeskIdleMinutes(input.legacyIdleRaw)
  }
  return DEFAULT_DESK_IDLE_MINUTES
}
