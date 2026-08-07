/** Team members shown as calendar columns / assignee options. */
export type CalendarTeamMember = {
  id: string
  fullName: string
  defaultSiteId?: string
  includeInCalendar?: boolean
}

/** Unwrap GET /api/vet/team payload (array or `{ members }`). */
export function extractTeamMembersList(res: unknown): unknown[] {
  if (Array.isArray(res)) return res
  if (res && typeof res === 'object') {
    const root = res as { data?: unknown; members?: unknown }
    const payload = root.data ?? root
    if (Array.isArray(payload)) return payload
    if (payload && typeof payload === 'object') {
      const members = (payload as { members?: unknown }).members
      if (Array.isArray(members)) return members
    }
  }
  return []
}

/**
 * Map API team rows to calendar assignees.
 * Excludes `includeInCalendar === false`; optional site filter mirrors defaultSiteId rules.
 */
export function mapTeamMembersForCalendar(raw: unknown[], siteId?: string): CalendarTeamMember[] {
  const sid = (siteId || '').trim()
  return raw
    .filter((row): row is Record<string, unknown> => {
      if (!row || typeof row !== 'object') return false
      return !!(row as Record<string, unknown>).userId
    })
    .map((m) => ({
      id: String(m.userId),
      fullName: String(m.fullName || m.email || m.userId),
      defaultSiteId: m.defaultSiteId ? String(m.defaultSiteId) : '',
      includeInCalendar: m.includeInCalendar !== false,
    }))
    .filter((m) => {
      if (!m.includeInCalendar) return false
      return !sid || !m.defaultSiteId || m.defaultSiteId === sid
    })
}
