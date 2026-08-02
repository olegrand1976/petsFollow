/**
 * Item de nav Pro (offre / ops) vers `/presentation`.
 */
export function presentationNavItem(
  label: string,
  section: string,
): {
  to: string
  label: string
  icon: 'slideshow'
  section: string
} {
  return { to: '/presentation', label, icon: 'slideshow', section }
}
