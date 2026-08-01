/**
 * Item de nav Pro (offre / ops) vers `/flux`.
 * Forme compatible `ProNavItem` (évite d’importer le SFC sidebar en tests unitaires).
 */
export function productFlowsNavItem(
  label: string,
  section: string,
): {
  to: string
  label: string
  icon: 'hub'
  section: string
} {
  return { to: '/flux', label, icon: 'hub', section }
}
