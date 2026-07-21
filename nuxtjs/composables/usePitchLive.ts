/**
 * Client audio full-duplex pour l'entraînement pitch (Gemini Live via WS Go).
 * - Capture micro → AudioWorklet → PCM16 16 kHz mono → frames binaires WS (~100 ms)
 * - Lecture PCM16 24 kHz reçu en binaire, file programmée + barge-in (flush)
 * - Mix micro + voix véto exposé pour l'enregistrement replay (MediaRecorder)
 */

export type PitchLiveCallbacks = {
  onReady: () => void
  onTranscript: (role: 'vet' | 'commercial', textDelta: string) => void
  onInterrupted: () => void
  onEnded: (outcome: string, appointmentSlot?: string, reason?: string) => void
  onClosed: () => void
}

const CAPTURE_RATE = 16000
const PLAYBACK_RATE = 24000
const CHUNK_SAMPLES = 1600 // 100 ms à 16 kHz

// Worklet inline : accumule des blocs de 100 ms et poste des Int16Array.
const workletSource = `
class PfPcmCapture extends AudioWorkletProcessor {
  constructor() {
    super()
    this.buf = new Int16Array(${CHUNK_SAMPLES})
    this.len = 0
  }
  process(inputs) {
    const ch = inputs[0] && inputs[0][0]
    if (!ch) return true
    for (let i = 0; i < ch.length; i++) {
      const s = Math.max(-1, Math.min(1, ch[i]))
      this.buf[this.len++] = s < 0 ? s * 0x8000 : s * 0x7fff
      if (this.len === this.buf.length) {
        this.port.postMessage(this.buf.slice(0))
        this.len = 0
      }
    }
    return true
  }
}
registerProcessor('pf-pcm-capture', PfPcmCapture)
`

function readCookie(name: string): string {
  if (typeof document === 'undefined') return ''
  const m = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return m ? decodeURIComponent(m[1]) : ''
}

export function usePitchLive() {
  // Capturé au setup — les appels connect() ont lieu hors contexte Nuxt (handlers).
  const apiBase = useRuntimeConfig().public.apiBase as string

  let ws: WebSocket | null = null
  let captureCtx: AudioContext | null = null
  let playbackCtx: AudioContext | null = null
  let micStream: MediaStream | null = null
  let workletNode: AudioWorkletNode | null = null
  let mixDest: MediaStreamAudioDestinationNode | null = null
  let nextPlayTime = 0
  let activeSources: AudioBufferSourceNode[] = []
  let ready = false

  /** Flux mixé micro + véto, disponible après connect() réussi (pour MediaRecorder). */
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
            await startCapture()
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

  async function startCapture() {
    if (!micStream) return
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
    src.connect(workletNode)
    // Worklet sans sortie audible — pas de connexion à destination.

    // Contexte de lecture + mix replay (micro + véto).
    playbackCtx = new AudioContext({ sampleRate: PLAYBACK_RATE })
    mixDest = playbackCtx.createMediaStreamDestination()
    playbackCtx.createMediaStreamSource(micStream).connect(mixDest)
    nextPlayTime = 0
    // Politique autoplay : reprendre les contextes (créés dans le geste utilisateur).
    void captureCtx.resume().catch(() => {})
    void playbackCtx.resume().catch(() => {})
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

  /** Barge-in : vide la file de lecture immédiatement. */
  function flushPlayback() {
    for (const s of activeSources) {
      try { s.stop() } catch { /* déjà stoppé */ }
    }
    activeSources = []
    nextPlayTime = 0
  }

  /** Envoi d'une ligne tapée (entrée dégradée pendant le live). */
  function sendText(text: string) {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'text', text }))
    }
  }

  /** Raccrocher manuellement : informe le serveur puis ferme. */
  function end() {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'end' }))
    }
    stop()
  }

  function stop() {
    ready = false
    flushPlayback()
    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.onmessage = null
      try { ws.close() } catch { /* ignore */ }
      ws = null
    }
    workletNode?.port.close()
    workletNode = null
    micStream?.getTracks().forEach(t => t.stop())
    micStream = null
    void captureCtx?.close().catch(() => {})
    captureCtx = null
    void playbackCtx?.close().catch(() => {})
    playbackCtx = null
    mixDest = null
  }

  return { connect, mixedStream, sendText, end, stop }
}
