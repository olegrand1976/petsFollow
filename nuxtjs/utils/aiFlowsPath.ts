/**
 * Item de nav Pro (section IA) vers `/flux-ia`.
 */
export function aiFlowsNavItem(
  label: string,
  section: string,
): {
  to: string
  label: string
  icon: 'account_tree'
  section: string
} {
  return { to: '/flux-ia', label, icon: 'account_tree', section }
}
