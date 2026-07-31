import { describe, expect, it } from 'vitest'
import { resolvePacsViewerEngine } from '../../utils/pacs-viewer-engine'

describe('resolvePacsViewerEngine', () => {
  it('defaults to canvas', () => {
    expect(resolvePacsViewerEngine(undefined)).toBe('canvas')
    expect(resolvePacsViewerEngine('')).toBe('canvas')
    expect(resolvePacsViewerEngine('png')).toBe('canvas')
  })

  it('accepts cornerstone aliases', () => {
    expect(resolvePacsViewerEngine('cornerstone')).toBe('cornerstone')
    expect(resolvePacsViewerEngine('CS3D')).toBe('cornerstone')
    expect(resolvePacsViewerEngine('true')).toBe('cornerstone')
  })
})
