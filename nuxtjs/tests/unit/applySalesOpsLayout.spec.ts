import { beforeEach, describe, expect, it, vi } from 'vitest'

const setPageLayout = vi.hoisted(() => vi.fn())
vi.stubGlobal('setPageLayout', setPageLayout)

import { applySalesOpsLayout } from '../../utils/applySalesOpsLayout'

describe('applySalesOpsLayout', () => {
  beforeEach(() => {
    setPageLayout.mockClear()
  })

  it('pose le layout pour admin / commercial / manager', () => {
    expect(applySalesOpsLayout('admin')).toBe(true)
    expect(setPageLayout).toHaveBeenCalledWith('admin')
    expect(applySalesOpsLayout('commercial')).toBe(true)
    expect(setPageLayout).toHaveBeenCalledWith('commercial')
    expect(applySalesOpsLayout('commercial_manager')).toBe(true)
    expect(setPageLayout).toHaveBeenCalledWith('commercial-manager')
  })

  it('ignore les autres rôles', () => {
    expect(applySalesOpsLayout('vet')).toBe(false)
    expect(applySalesOpsLayout(null)).toBe(false)
    expect(setPageLayout).not.toHaveBeenCalled()
  })
})
