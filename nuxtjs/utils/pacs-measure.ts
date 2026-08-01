/**
 * PACS length calibration helpers (G4 / P2.2).
 * DICOM PixelSpacing is row\\col → [rowMmPerPx, colMmPerPx].
 */

export type PacsSpacingMm = readonly [number, number]

export type PacsLengthResult = {
  length: number
  unit: 'mm' | 'px'
  calibrated: boolean
}

export type PacsMatrixSize = {
  rows: number
  columns: number
}

export type PacsSegment = {
  x1: number
  y1: number
  x2: number
  y2: number
}

/** Parse API `pixelSpacingMm` ([row, col]) or reject invalid values. */
export function parsePixelSpacingMm(raw: unknown): PacsSpacingMm | null {
  if (!Array.isArray(raw) || raw.length < 2) return null
  const row = Number(raw[0])
  const col = Number(raw[1])
  if (!Number.isFinite(row) || !Number.isFinite(col) || row <= 0 || col <= 0) return null
  return [row, col]
}

export function parseMatrixSize(rows: unknown, columns: unknown): PacsMatrixSize | null {
  const r = Number(rows)
  const c = Number(columns)
  if (!Number.isFinite(r) || !Number.isFinite(c) || r < 1 || c < 1) return null
  return { rows: Math.round(r), columns: Math.round(c) }
}

/**
 * Preview PNG/JPEG may be resized by Orthanc. Only trust PixelSpacing for canvas
 * measures when the loaded bitmap matches DICOM Rows×Columns (±1 px tolerance).
 */
export function spacingUsableForBitmap(
  spacing: PacsSpacingMm | null,
  matrix: PacsMatrixSize | null,
  bitmapWidth: number,
  bitmapHeight: number,
): PacsSpacingMm | null {
  if (!spacing || !matrix) return null
  if (bitmapWidth < 1 || bitmapHeight < 1) return null
  const wOk = Math.abs(bitmapWidth - matrix.columns) <= 1
  const hOk = Math.abs(bitmapHeight - matrix.rows) <= 1
  return wOk && hOk ? spacing : null
}

/**
 * Scale DICOM PixelSpacing to Orthanc preview bitmap size when the PNG is
 * downsampled (approx. mm still consistent with on-screen anatomy).
 */
export function scaleSpacingToBitmap(
  spacing: PacsSpacingMm | null,
  matrix: PacsMatrixSize | null,
  bitmapWidth: number,
  bitmapHeight: number,
): PacsSpacingMm | null {
  if (!spacing || !matrix) return null
  if (bitmapWidth < 1 || bitmapHeight < 1) return null
  const rowScale = matrix.rows / bitmapHeight
  const colScale = matrix.columns / bitmapWidth
  if (!Number.isFinite(rowScale) || !Number.isFinite(colScale) || rowScale <= 0 || colScale <= 0) {
    return null
  }
  return [spacing[0] * rowScale, spacing[1] * colScale]
}

/**
 * Convert a screen-space segment (canvas coords after pan/zoom paint) to image length.
 * `scale` is the viewer zoom factor (image pixels × scale = screen pixels).
 */
export function measureScreenSegment(
  dxScreen: number,
  dyScreen: number,
  scale: number,
  spacing: PacsSpacingMm | null,
): PacsLengthResult {
  const s = scale > 0 && Number.isFinite(scale) ? scale : 1
  return measureImageSegment(dxScreen / s, dyScreen / s, spacing)
}

/** Length from image-space deltas (origin = image centre, units = bitmap px). */
export function measureImageSegment(
  dxImg: number,
  dyImg: number,
  spacing: PacsSpacingMm | null,
): PacsLengthResult {
  if (spacing) {
    const [rowMm, colMm] = spacing
    const mm = Math.hypot(dxImg * colMm, dyImg * rowMm)
    return { length: mm, unit: 'mm', calibrated: true }
  }
  return { length: Math.hypot(dxImg, dyImg), unit: 'px', calibrated: false }
}

export function formatPacsLength(result: PacsLengthResult, digits = 1): string {
  const n = result.unit === 'mm'
    ? result.length.toFixed(digits)
    : String(Math.round(result.length))
  return `${n} ${result.unit}`
}

/**
 * Canvas paint uses: translate(w/2+offsetX, h/2+offsetY) then scale(s).
 * Image coords: origin at image centre, +x right, +y down (bitmap pixels).
 */
export function screenToImage(
  screenX: number,
  screenY: number,
  canvasW: number,
  canvasH: number,
  offsetX: number,
  offsetY: number,
  scale: number,
): { x: number, y: number } {
  const s = scale > 0 && Number.isFinite(scale) ? scale : 1
  return {
    x: (screenX - canvasW / 2 - offsetX) / s,
    y: (screenY - canvasH / 2 - offsetY) / s,
  }
}

export function imageToScreen(
  imgX: number,
  imgY: number,
  canvasW: number,
  canvasH: number,
  offsetX: number,
  offsetY: number,
  scale: number,
): { x: number, y: number } {
  const s = scale > 0 && Number.isFinite(scale) ? scale : 1
  return {
    x: imgX * s + canvasW / 2 + offsetX,
    y: imgY * s + canvasH / 2 + offsetY,
  }
}

export function segmentImageToScreen(
  seg: PacsSegment,
  canvasW: number,
  canvasH: number,
  offsetX: number,
  offsetY: number,
  scale: number,
): PacsSegment {
  const a = imageToScreen(seg.x1, seg.y1, canvasW, canvasH, offsetX, offsetY, scale)
  const b = imageToScreen(seg.x2, seg.y2, canvasW, canvasH, offsetX, offsetY, scale)
  return { x1: a.x, y1: a.y, x2: b.x, y2: b.y }
}
