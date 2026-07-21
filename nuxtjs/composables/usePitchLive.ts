/**
 * Client audio full-duplex pour l'entraînement pitch (Gemini Live via WS Go).
 * - Capture micro → AudioWorklet → downsample PCM16 16 kHz mono → frames binaires WS (~100 ms)
 * - Lecture PCM16 24 kHz reçu en binaire, file programmée + barge-in (flush)
 * - Mix micro + voix véto exposé pour l'enregistrement replay (MediaRecorder)
 */

export type PitchLiveCallbacks = {
  onReady: () => void
  onTranscript: (role: 'vet' | 'commercial', text: string) => void
  onInterrupted: () => void
  /** Fin d'un tour modèle : le prochain transcript du même rôle ouvre une nouvelle ligne. */
  onTurnComplete?: () => void
  onEnded: (outcome: string, appointmentSlot?: string, reason?: string) => void
  onClosed: () => void
}

const CAPTURE_RATE = 16000
const PLAYBACK_RATE = 24000
const CHUNK_SAMPLES = 1600 // 100 ms à 16 kHz

// Worklet : downsample linéaire vers 16 kHz (sampleRate navigateur souvent 48 kHz).
const workletSource = `
class PfPcmCapture extends AudioWorkletProcessor {
  constructor() {
    super()
    this.targetRate = ${CAPTURE_RATE}
    this.ratio = sampleRate / this.targetRate
    this.buf = new Int16Array(${CHUNK_SAMPLES})
    this.len = 0
    this.phase = 0
  }
  process(inputs) {
    const ch = inputs[0] && inputs[0][0]
    if (!ch) return true
    while (this.phase < ch.length) {
      const i0 = Math.floor(this.phase)
      const i1 = Math.min(i0 + 1, ch.length - 1)
      const frac = this.phase - i0
      const f = ch[i0] * (1 - frac) + ch[i1] * frac
      const s = Math.max(-1, Math.min(1, f))
      this.buf[this.len++] = s < 0 ? s * 0x8000 : s * 0x7fff
      if (this.len === this.buf.length) {
        this.port.postMessage(this.buf.slice(0))
        this.len = 0
      }
      this.phase += this.ratio
    }
    this.phase -= ch.length
    return true
  }
}
registerProcessor('pf-pcm-capture', PfPcmCapture)
`

/** Fusionne un fragment Gemini (delta ou snapshot cumulatif) dans le texte courant. */
export function mergeTranscriptChunk(current: string, incoming: string): string {
  const next = incoming ?? ''
  if (!next) return current
  if (!current) return next
  if (next.startsWith(current)) return next
  if (current.startsWith(next)) return current
  // Extension : current est préfixe d'un début de next (overlap partiel rare).
  for (let n = Math.min(current.length, next.length); n > 0; n--) {
    if (current.endsWith(next.slice(0, n))) {
      return current + next.slice(n)
    }
  }
  const needSpace = !/\s$/.test(current) && !/^\s/.test(next)
  return current + (needSpace ? ' ' : '') + next
}

function readCookie(name: string): string {
  if (typeof document === 'undefined') return ''
  const m = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return m ? decodeURIComponent(m[1]) : ''
}

type BrowserAudioContext = typeof AudioContext

function getAudioContextCtor(): BrowserAudioContext | null {
  if (typeof window === 'undefined') return null
  return window.AudioContext || (window as unknown as { webkitAudioContext?: BrowserAudioContext }).webkitAudioContext || null
}

/**
 * À appeler de façon synchrone dans le handler du clic « Appeler ».
 * Débloque l'AudioContext avant tout await ($fetch / WS), sinon la sonnerie est muette (autoplay policy).
 */
export function unlockPhoneAudio(): AudioContext | null {
  const Ctor = getAudioContextCtor()
  if (!Ctor) return null
  const ctx = new Ctor()
  // Buffer silencieux : certains navigateurs n'autorisent le son qu'après un start() dans le geste user.
  try {
    const buf = ctx.createBuffer(1, 1, ctx.sampleRate || 22050)
    const src = ctx.createBufferSource()
    src.buffer = buf
    src.connect(ctx.destination)
    src.start(0)
  } catch {
    /* ignore */
  }
  void ctx.resume()
  return ctx
}

/** Sonnerie téléphone réaliste (~2,5 s) via oscillateurs Web Audio. */
export async function playPhoneRingtone(durationMs = 2500, existingCtx?: AudioContext | null): Promise<void> {
  if (typeof window === 'undefined') {
    await new Promise(r => setTimeout(r, durationMs))
    return
  }
  const owned = !existingCtx
  const Ctor = getAudioContextCtor()
  const ctx = existingCtx ?? (Ctor ? new Ctor() : null)
  if (!ctx) {
    await new Promise(r => setTimeout(r, durationMs))
    return
  }
  try {
    if (ctx.state === 'suspended') {
      await ctx.resume()
    }
    if (ctx.state !== 'running') {
      // Toujours attendre la durée visuelle même si l'audio reste bloqué.
      await new Promise(r => setTimeout(r, durationMs))
      return
    }
    const master = ctx.createGain()
    master.gain.value = 0.35
    master.connect(ctx.destination)
    const t0 = ctx.currentTime
    // Deux doublets type sonnerie européenne (440/480 Hz).
    const bursts: Array<[number, number]> = [
      [0.0, 0.4],
      [0.5, 0.9],
      [1.4, 1.8],
      [1.9, 2.3],
    ]
    for (const [start, end] of bursts) {
      for (const freq of [440, 480]) {
        const osc = ctx.createOscillator()
        const g = ctx.createGain()
        osc.type = 'sine'
        osc.frequency.value = freq
        g.gain.setValueAtTime(0, t0 + start)
        g.gain.linearRampToValueAtTime(0.45, t0 + start + 0.02)
        g.gain.setValueAtTime(0.45, t0 + end - 0.04)
        g.gain.linearRampToValueAtTime(0, t0 + end)
        osc.connect(g)
        g.connect(master)
        osc.start(t0 + start)
        osc.stop(t0 + end + 0.01)
      }
    }
    await new Promise(r => setTimeout(r, durationMs))
  } finally {
    if (owned) {
      void ctx.close().catch(() => {})
    }
  }
}

export function usePitchLive() {
  // Capturé au setup — les appels connect() ont lieu hors contexte Nuxt (handlers).
  const apiBase = useRuntimeConfig().public.apiBase as string

  let ws: WebSocket | null = null
  let captureCtx: AudioContext | null = null
  let playbackCtx: AudioContext | null = null
  let micStream: MediaStream | null = null
  let workletNode: AudioWorkletNode | null = null
  let silentGain: GainNode | null = null
  let mixDest: MediaStreamAudioDestinationNode | null = null
  let nextPlayTime = 0
  let activeSources: AudioBufferSourceNode[] = []
  let ready = false
  let captureStarted = false

  /** Flux mixé micro + véto, disponible après startOpening() (pour MediaRecorder). */
  function mixedStream(): MediaStream | null {
    return mixDest?.stream ?? null
  }

  async function connect(simId: string, cb: PitchLiveCallbacks): Promise<boolean> {
    const token = readCookie('pf_token')
    if (!token) return false
    const wsUrl = apiBase.replace(/^http/, 'ws')
      + `/api/v1/commercial/pitch-sims/${simId}/live?token=${encodeURIComponent(token)}`

    try {
      micStream = await navigator.mediaDevices.getUserMedia({
        audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true },
      })
    } catch {
      return false
    }

    const opened = await new Promise<boolean>((resolve) => {
      try {
        ws = new WebSocket(wsUrl)
      } catch {
        resolve(false)
        return
      }
      ws.binaryType = 'arraybuffer'
      const failTimer = setTimeout(() => resolve(false), 8000)

      ws.onmessage = async (ev) => {
        if (ev.data instanceof ArrayBuffer) {
          playPcm(new Int16Array(ev.data))
          return
        }
        let msg: any
        try { msg = JSON.parse(ev.data) } catch { return }
        switch (msg.type) {
          case 'ready':
            clearTimeout(failTimer)
            ready = true
            // Micro + Allo différés : le client joue d'abord la sonnerie.
            cb.onReady()
            resolve(true)
            break
          case 'transcript':
            cb.onTranscript(msg.role, msg.text)
            break
          case 'interrupted':
            flushPlayback()
            cb.onInterrupted()
            break
          case 'turn_complete':
            cb.onTurnComplete?.()
            break
          case 'ended':
            cb.onEnded(msg.outcome, msg.appointmentSlot, msg.reason)
            break
          case 'error':
            clearTimeout(failTimer)
            resolve(false)
            break
        }
      }
      ws.onerror = () => {
        clearTimeout(failTimer)
        resolve(false)
      }
      ws.onclose = () => {
        clearTimeout(failTimer)
        if (!ready) resolve(false)
        cb.onClosed()
      }
    })

    if (!opened) {
      stop()
      return false
    }
    return true
  }

  /** Après la sonnerie : ouvre le micro/lecture puis déclenche l'Allo (évite de dropper l'audio). */
  async function startOpening() {
    if (!ready || ws?.readyState !== WebSocket.OPEN) return
    const ok = await startCapture()
    if (ok && playbackCtx && ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'start_opening' }))
    }
  }

  async function startCapture(): Promise<boolean> {
    if (!micStream) return false
    if (captureStarted && playbackCtx) return true
    // Échec partiel précédent : autorise un retry propre.
    if (captureStarted && !playbackCtx) {
      captureStarted = false
    }
    captureStarted = true
    try {
      captureCtx = new AudioContext({ sampleRate: CAPTURE_RATE })
      const blobUrl = URL.createObjectURL(new Blob([workletSource], { type: 'application/javascript' }))
      try {
        await captureCtx.audioWorklet.addModule(blobUrl)
      } finally {
        URL.revokeObjectURL(blobUrl)
      }
      const src = captureCtx.createMediaStreamSource(micStream)
      workletNode = new AudioWorkletNode(captureCtx, 'pf-pcm-capture')
      workletNode.port.onmessage = (e: MessageEvent<Int16Array>) => {
        if (ws?.readyState === WebSocket.OPEN) {
          ws.send(e.data.buffer)
        }
      }
      // Keep-alive : certains navigateurs n'exécutent process() que si branché à destination.
      silentGain = captureCtx.createGain()
      silentGain.gain.value = 0
      src.connect(workletNode)
      workletNode.connect(silentGain)
      silentGain.connect(captureCtx.destination)

      playbackCtx = new AudioContext({ sampleRate: PLAYBACK_RATE })
      mixDest = playbackCtx.createMediaStreamDestination()
      playbackCtx.createMediaStreamSource(micStream).connect(mixDest)
      nextPlayTime = 0
      await Promise.all([
        captureCtx.resume().catch(() => {}),
        playbackCtx.resume().catch(() => {}),
      ])
      return true
    } catch {
      captureStarted = false
      try { workletNode?.port.close() } catch { /* ignore */ }
      workletNode = null
      silentGain = null
      void captureCtx?.close().catch(() => {})
      captureCtx = null
      void playbackCtx?.close().catch(() => {})
      playbackCtx = null
      mixDest = null
      return false
    }
  }

  function playPcm(pcm: Int16Array) {
    if (!playbackCtx || pcm.length === 0) return
    const floats = new Float32Array(pcm.length)
    for (let i = 0; i < pcm.length; i++) floats[i] = pcm[i] / 0x8000
    const buf = playbackCtx.createBuffer(1, floats.length, PLAYBACK_RATE)
    buf.copyToChannel(floats, 0)
    const srcNode = playbackCtx.createBufferSource()
    srcNode.buffer = buf
    srcNode.connect(playbackCtx.destination)
    if (mixDest) srcNode.connect(mixDest)
    const startAt = Math.max(playbackCtx.currentTime, nextPlayTime)
    srcNode.start(startAt)
    nextPlayTime = startAt + buf.duration
    activeSources.push(srcNode)
    srcNode.onended = () => {
      activeSources = activeSources.filter(s => s !== srcNode)
    }
  }

  function flushPlayback() {
    for (const s of activeSources) {
      try { s.stop() } catch { /* déjà stoppé */ }
    }
    activeSources = []
    nextPlayTime = 0
  }

  function sendText(text: string) {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'text', text }))
    }
  }

  function end() {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'end' }))
    }
    stop()
  }

  function stop() {
    ready = false
    captureStarted = false
    flushPlayback()
    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.onmessage = null
      try { ws.close() } catch { /* ignore */ }
      ws = null
    }
    try { workletNode?.port.close() } catch { /* ignore */ }
    workletNode = null
    silentGain = null
    micStream?.getTracks().forEach(t => t.stop())
    micStream = null
    void captureCtx?.close().catch(() => {})
    captureCtx = null
    void playbackCtx?.close().catch(() => {})
    playbackCtx = null
    mixDest = null
  }

  return { connect, startOpening, mixedStream, sendText, end, stop }
}
