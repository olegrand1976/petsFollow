/** Pure helpers for practice header quick-links prefs (settings + Vitest). */

export type HeaderLinkCustom = { id: string, label: string, url: string }
export type HeaderLinkCatalogRow = { id: string, label: string, url: string, enabled: boolean }
export type HeaderLinksPrefs = {
  configured: boolean
  enabled: string[]
  order: string[]
  custom: HeaderLinkCustom[]
}

export type HeaderLinkUnifiedRow =
  | { kind: 'catalog', id: string, label: string, url: string, enabled: boolean }
  | { kind: 'custom', id: string, label: string, url: string }

export function isValidHttpsUrl(raw: string): boolean {
  const s = raw.trim()
  if (!s.startsWith('https://') || s.length > 2048) return false
  try {
    const u = new URL(s)
    return u.protocol === 'https:' && !!u.host && !u.username && !u.password
  } catch {
    return false
  }
}

export function isCompleteCustom(c: { label: string, url: string }): boolean {
  const label = c.label.trim()
  return label.length > 0 && label.length <= 40 && isValidHttpsUrl(c.url)
}

/** Returns an i18n key suffix under headerLinks.* or null if OK. */
export function validateHeaderLinksCustoms(customs: HeaderLinkCustom[]): string | null {
  const seen = new Set<string>()
  for (const c of customs) {
    const label = c.label.trim()
    const url = c.url.trim()
    // Draft rows left empty after "Add" — treat as incomplete if either field touched.
    if (!label && (url === '' || url === 'https://')) {
      return 'incompleteCustom'
    }
    if (!label || label.length > 40) return 'incompleteCustom'
    if (!isValidHttpsUrl(url)) return 'urlInvalid'
    const key = url.toLowerCase()
    if (seen.has(key)) return 'duplicateUrl'
    seen.add(key)
  }
  return null
}

export function buildUnifiedRows(
  catalog: HeaderLinkCatalogRow[],
  prefs: HeaderLinksPrefs,
): HeaderLinkUnifiedRow[] {
  const customById = new Map((prefs.custom || []).map((c) => [c.id, c]))
  const catalogById = new Map((catalog || []).map((c) => [c.id, c]))
  const rows: HeaderLinkUnifiedRow[] = []
  const seen = new Set<string>()

  const pushCatalog = (id: string) => {
    const c = catalogById.get(id)
    if (!c || seen.has(id)) return
    seen.add(id)
    rows.push({ kind: 'catalog', id: c.id, label: c.label, url: c.url, enabled: c.enabled })
  }
  const pushCustom = (id: string) => {
    const c = customById.get(id)
    if (!c || seen.has(id)) return
    seen.add(id)
    rows.push({ kind: 'custom', id: c.id, label: c.label, url: c.url })
  }

  for (const id of prefs.order || []) {
    if (id.startsWith('custom_')) pushCustom(id)
    else pushCatalog(id)
  }
  for (const c of catalog || []) pushCatalog(c.id)
  for (const c of prefs.custom || []) pushCustom(c.id)
  return rows
}

export function prefsFromUnifiedRows(rows: HeaderLinkUnifiedRow[]): HeaderLinksPrefs {
  const enabled: string[] = []
  const order: string[] = []
  const custom: HeaderLinkCustom[] = []

  for (const row of rows) {
    if (row.kind === 'catalog') {
      if (row.enabled) {
        enabled.push(row.id)
        order.push(row.id)
      }
      continue
    }
    custom.push({ id: row.id, label: row.label.trim(), url: row.url.trim() })
    enabled.push(row.id)
    order.push(row.id)
  }
  return { configured: true, enabled, order, custom }
}

export function moveRow<T>(rows: T[], idx: number, delta: number): T[] {
  const next = idx + delta
  if (next < 0 || next >= rows.length) return rows
  const arr = [...rows]
  const [row] = arr.splice(idx, 1)
  arr.splice(next, 0, row)
  return arr
}
