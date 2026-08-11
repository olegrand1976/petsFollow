/** Support ticket status workflow (aligned with go/internal/store/support_status.go). */

export const SUPPORT_STATUSES = ['open', 'in_progress', 'to_test', 'done', 'closed'] as const

export type SupportStatus = (typeof SUPPORT_STATUSES)[number]

const TRANSITIONS: Record<SupportStatus, SupportStatus[]> = {
  open: ['in_progress', 'closed'],
  in_progress: ['to_test', 'open', 'closed'],
  to_test: ['done', 'in_progress', 'closed'],
  done: ['closed'],
  closed: ['open'],
}

export function isSupportStatus (raw: string | null | undefined): raw is SupportStatus {
  return SUPPORT_STATUSES.includes(raw as SupportStatus)
}

export function isValidSupportStatusTransition (from: string, to: string): boolean {
  if (!isSupportStatus(from) || !isSupportStatus(to)) return false
  if (from === to) return true
  return TRANSITIONS[from].includes(to)
}

/** Options for the status select: current + allowed next. */
export function supportStatusSelectOptions (current: string): SupportStatus[] {
  if (!isSupportStatus(current)) return [...SUPPORT_STATUSES]
  const next = TRANSITIONS[current]
  return [current, ...next.filter(s => s !== current)]
}
