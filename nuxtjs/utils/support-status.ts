/** Support ticket status workflow (aligned with go/internal/store/support_status.go). */

export const SUPPORT_STATUSES = ['open', 'in_progress', 'to_test', 'done', 'closed'] as const

export type SupportStatus = (typeof SUPPORT_STATUSES)[number]

/** Display order for workflow stepper (closed = abandon / terminal). */
export const SUPPORT_WORKFLOW_ORDER: readonly SupportStatus[] = SUPPORT_STATUSES

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

/** True when `to` is a different status reachable from `from`. */
export function canAdvanceSupportStatus (from: string, to: string): boolean {
  if (!isSupportStatus(from) || !isSupportStatus(to) || from === to) return false
  return TRANSITIONS[from].includes(to)
}

export function supportStatusBadgeVariant (
  status: string,
): 'neutral' | 'success' | 'warning' | 'danger' {
  switch (status as SupportStatus | string) {
    case 'open':
      return 'danger'
    case 'in_progress':
      return 'warning'
    case 'to_test':
      return 'neutral'
    case 'done':
      return 'success'
    case 'closed':
      return 'neutral'
    default:
      return 'neutral'
  }
}
