/** First usable given name for dashboard greeting — skips titles like Dr / Mme. */
const TITLE_PREFIXES = new Set([
  'dr',
  'dr.',
  'mme',
  'mme.',
  'mr',
  'mr.',
  'mrs',
  'mrs.',
  'ms',
  'ms.',
  'm.',
  'mlle',
  'mlle.',
  'prof',
  'prof.',
  'pr',
  'pr.',
])

export function welcomeDisplayName(fullName?: string | null): string {
  if (!fullName) return ''
  const parts = fullName.trim().split(/\s+/).filter(Boolean)
  while (parts.length > 1 && TITLE_PREFIXES.has(parts[0].toLowerCase())) {
    parts.shift()
  }
  return parts[0] || ''
}
