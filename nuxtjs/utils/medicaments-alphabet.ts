/** Dictionary letters for national AFMPS medication browse (A–Z + non-letter bucket). */
export const MEDICAMENT_ALPHABET_LETTERS = [
  ...'ABCDEFGHIJKLMNOPQRSTUVWXYZ'.split(''),
  '#',
] as const

export type MedicamentLetter = (typeof MEDICAMENT_ALPHABET_LETTERS)[number]

export function normalizeMedicamentLetter(raw: string | null | undefined): MedicamentLetter | null {
  const s = (raw ?? '').trim().toUpperCase()
  if (s === '#') return '#'
  if (s.length === 1 && s >= 'A' && s <= 'Z') return s as MedicamentLetter
  return null
}

/** Ensure every A–Z/# key exists (missing → 0). */
export function parseLetterCounts(raw: Record<string, unknown> | null | undefined): Record<MedicamentLetter, number> {
  const out = {} as Record<MedicamentLetter, number>
  for (const letter of MEDICAMENT_ALPHABET_LETTERS) {
    const n = raw?.[letter]
    out[letter] = typeof n === 'number' && Number.isFinite(n) ? n : Number(n) || 0
  }
  return out
}

export function letterHasEntries(counts: Record<MedicamentLetter, number>, letter: MedicamentLetter): boolean {
  return (counts[letter] ?? 0) > 0
}

export type AfmpsMetaFields = {
  source?: string
  manufacturer?: string
  activeSubstance?: string
  strength?: string
}

/** Normalize AFMPS / Compendium meta keys (snake or camel). */
export function parseAfmpsMeta(raw: unknown): AfmpsMetaFields {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return {}
  const m = raw as Record<string, unknown>
  const str = (v: unknown) => (typeof v === 'string' && v.trim() ? v.trim() : undefined)
  return {
    source: str(m.source),
    manufacturer: str(m.manufacturer),
    activeSubstance: str(m.active_substance) ?? str(m.activeSubstance),
    strength: str(m.strength),
  }
}

export function afmpsSourceKind(source: string | null | undefined): 'afmps' | 'compendium' | 'other' | null {
  if (!source) return null
  const s = source.toLowerCase()
  if (s.includes('afmps')) return 'afmps'
  if (s.includes('compendium')) return 'compendium'
  return 'other'
}
