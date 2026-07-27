/**
 * Base path pages use cases (staging) — une URL pour admin / commercial / manager.
 */
export function usecasesBasePath(_role?: string | null): string {
  return '/usecases'
}

/** Item de nav Pro (staging|local uniquement). */
export function usecasesNavItem(label: string, section: string): {
  to: string
  label: string
  icon: 'checklist'
  section: string
} {
  return { to: '/usecases', label, icon: 'checklist', section }
}
