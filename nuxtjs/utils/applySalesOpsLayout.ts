/**
 * Pose le shell Pro pour admin / commercial / commercial_manager.
 * @returns true si un layout sales/ops a été appliqué.
 */
export function applySalesOpsLayout(role: string | null | undefined): boolean {
  if (role === 'admin') {
    setPageLayout('admin')
    return true
  }
  if (role === 'commercial') {
    setPageLayout('commercial')
    return true
  }
  if (role === 'commercial_manager') {
    setPageLayout('commercial-manager')
    return true
  }
  return false
}
