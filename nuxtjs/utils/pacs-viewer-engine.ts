export type PacsViewerEngine = 'canvas' | 'cornerstone'

/** Default remains canvas until Cornerstone GA path is proven in staging. */
export function resolvePacsViewerEngine(raw: unknown): PacsViewerEngine {
  const v = String(raw ?? '').trim().toLowerCase()
  if (v === 'cornerstone' || v === 'cs3d' || v === '1' || v === 'true') return 'cornerstone'
  return 'canvas'
}
