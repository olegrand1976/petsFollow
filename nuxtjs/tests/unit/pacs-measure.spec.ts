import { describe, expect, it } from 'vitest'
import {
  formatPacsLength,
  imageToScreen,
  measureImageSegment,
  measureScreenSegment,
  parseMatrixSize,
  parsePixelSpacingMm,
  scaleSpacingToBitmap,
  screenToImage,
  segmentImageToScreen,
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

  it('demo-rx spacing 0.5 mm: 10 image px → 5.0 mm', () => {
    // Matches DEMO_PIXEL_SPACING_MM in scripts/gen-minimal-dicom.py
    const spacing = parsePixelSpacingMm([0.5, 0.5])!
    const matrix = parseMatrixSize(64, 64)!
    const usable = spacingUsableForBitmap(spacing, matrix, 64, 64)
    expect(usable).toEqual([0.5, 0.5])
    const r = measureImageSegment(10, 0, usable)
    expect(r).toMatchObject({ calibrated: true, unit: 'mm' })
    expect(r.length).toBeCloseTo(5, 5)
    expect(formatPacsLength(r)).toBe('5.0 mm')
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

  it('scales spacing when preview bitmap is downsampled', () => {
    const spacing = parsePixelSpacingMm([0.5, 0.5])!
    const matrix = parseMatrixSize(512, 512)!
    const scaled = scaleSpacingToBitmap(spacing, matrix, 256, 256)
    expect(scaled).toEqual([1, 1])
    const len = measureImageSegment(100, 0, scaled)
    expect(len.length).toBeCloseTo(100, 5)
    expect(len.unit).toBe('mm')
  })

  it('round-trips screen ↔ image coords under pan/zoom', () => {
    const canvasW = 800
    const canvasH = 600
    const offsetX = 40
    const offsetY = -20
    const scale = 2
    const img = { x: 30, y: -15 }
    const scr = imageToScreen(img.x, img.y, canvasW, canvasH, offsetX, offsetY, scale)
    const back = screenToImage(scr.x, scr.y, canvasW, canvasH, offsetX, offsetY, scale)
    expect(back.x).toBeCloseTo(img.x, 8)
    expect(back.y).toBeCloseTo(img.y, 8)
  })

  it('projects image segments to screen for paint', () => {
    const seg = segmentImageToScreen(
      { x1: 0, y1: 0, x2: 10, y2: 0 },
      400,
      300,
      0,
      0,
      2,
    )
    expect(seg.x1).toBeCloseTo(200, 5)
    expect(seg.y1).toBeCloseTo(150, 5)
    expect(seg.x2).toBeCloseTo(220, 5)
    expect(seg.y2).toBeCloseTo(150, 5)
  })
})
