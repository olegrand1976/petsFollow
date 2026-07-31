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
  const dxImg = dxScreen / s
  const dyImg = dyScreen / s
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
