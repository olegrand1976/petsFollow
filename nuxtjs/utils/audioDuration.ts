/** Formats seconds as MM:SS (or H:MM:SS if ≥ 1h). */
export function formatAudioClock(totalSec: number): string {
  const n = Math.max(0, Math.floor(Number(totalSec) || 0))
  const h = Math.floor(n / 3600)
  const m = Math.floor((n % 3600) / 60)
  const s = n % 60
  const mm = m.toString().padStart(2, '0')
  const ss = s.toString().padStart(2, '0')
  if (h > 0) return `${h}:${mm}:${ss}`
  return `${mm}:${ss}`
}

/** Best-effort duration from a Blob/File via HTMLAudioElement metadata. */
export function probeAudioDurationSec(file: Blob): Promise<number> {
  return new Promise((resolve) => {
    const url = URL.createObjectURL(file)
    const audio = new Audio()
    let settled = false
    const finish = (sec: number) => {
      if (settled) return
      settled = true
      URL.revokeObjectURL(url)
      resolve(sec > 0 && Number.isFinite(sec) ? Math.round(sec) : 0)
    }
    audio.preload = 'metadata'
    audio.onloadedmetadata = () => {
      const d = audio.duration
      finish(Number.isFinite(d) ? d : 0)
    }
    audio.onerror = () => finish(0)
    setTimeout(() => finish(0), 3000)
    audio.src = url
  })
}
