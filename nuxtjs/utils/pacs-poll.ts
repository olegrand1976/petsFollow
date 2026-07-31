export type PacsState = 'offline' | 'starting' | 'ready'

/** Polling interval (ms) for adaptive PACS status refresh. */
export function pacsPollIntervalMs(state: PacsState): number {
  if (state === 'starting') return 2000
  if (state === 'ready') return 30000
  return 10000
}
