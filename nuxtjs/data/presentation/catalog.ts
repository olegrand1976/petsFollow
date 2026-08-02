/**
 * Parcours présentation cabinet (admin / commercial / manager).
 * Libellés via i18n `presentation.*`.
 */
export const PRESENTATION_STEPS = [
  'welcome',
  'pain',
  'surfaces',
  'vetpro',
  'care_loop',
  'ai_in_app',
  'ai_automation',
  'ecosystem',
  'offer',
  'close',
] as const

export type PresentationStepId = (typeof PRESENTATION_STEPS)[number]

const stepIdSet = new Set<string>(PRESENTATION_STEPS)

export function parsePresentationStepId(raw: unknown): PresentationStepId | null {
  const value = Array.isArray(raw) ? raw[0] : raw
  if (typeof value !== 'string' || !value) return null
  return stepIdSet.has(value) ? (value as PresentationStepId) : null
}

export function presentationStepIndex(id: PresentationStepId): number {
  return PRESENTATION_STEPS.indexOf(id)
}
