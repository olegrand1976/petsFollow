import { describe, expect, it } from 'vitest'
import {
  formatPacsLength,
  measureScreenSegment,
  parseMatrixSize,
  parsePixelSpacingMm,
  spacingUsableForBitmap,
} from '../../utils/pacs-measure'

describe('pacs-measure', () => {
  it('parsePixelSpacingMm accepts row/col pairs', () => {
    expect(parsePixelSpacingMm([0.5, 0.5])).toEqual([0.5, 0.5])
    expect(parsePixelSpacingMm(['0.2', '0.3'])).toEqual([0.2, 0.3])
    expect(parsePixelSpacingMm([0, 0.5])).toBeNull()
    expect(parsePixelSpacingMm([0.5])).toBeNull()
    expect(parsePixelSpacingMm(null)).toBeNull()
  })

  it('calibrates isotropic spacing to mm', () => {
    const r = measureScreenSegment(100, 0, 1, [0.5, 0.5])
    expect(r.calibrated).toBe(true)
    expect(r.unit).toBe('mm')
    expect(r.length).toBeCloseTo(50, 5)
    expect(formatPacsLength(r)).toBe('50.0 mm')
  })

  it('accounts for viewer zoom scale', () => {
    const r = measureScreenSegment(100, 0, 2, [0.5, 0.5])
    expect(r.length).toBeCloseTo(25, 5)
  })

  it('uses anisotropic row/col spacing', () => {
    const r = measureScreenSegment(10, 10, 1, [0.4, 0.2])
    expect(r.length).toBeCloseTo(Math.hypot(2, 4), 5)
  })

  it('falls back to image pixels when spacing missing', () => {
    const r = measureScreenSegment(30, 40, 1, null)
    expect(r).toEqual({ length: 50, unit: 'px', calibrated: false })
    expect(formatPacsLength(r)).toBe('50 px')
  })

  it('gates canvas spacing on DICOM matrix vs bitmap size', () => {
    const spacing = parsePixelSpacingMm([0.5, 0.5])!
    const matrix = parseMatrixSize(512, 512)!
    expect(spacingUsableForBitmap(spacing, matrix, 512, 512)).toEqual(spacing)
    expect(spacingUsableForBitmap(spacing, matrix, 511, 512)).toEqual(spacing) // ±1
    expect(spacingUsableForBitmap(spacing, matrix, 256, 256)).toBeNull()
    expect(spacingUsableForBitmap(spacing, null, 512, 512)).toBeNull()
  })
})
