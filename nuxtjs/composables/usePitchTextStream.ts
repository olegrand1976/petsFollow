/**
 * Client WebSocket texte (fallback pitch) : deltas token → affichage progressif.
 * Distinct de usePitchLive (audio Gemini Live).
 */

export type PitchTextStreamCallbacks = {
  onReady: () => void
  onDelta: (delta: string) => void
  onTurnComplete: (payload: {
    reply: string
    action: string
    ended: boolean
    outcome: string
    appointmentSlot?: string
    reason?: string
    interrupted?: boolean
  }) => void
  onInterrupted: () => void
  onEnded: (outcome: string, appointmentSlot?: string) => void
  onClosed: () => void
  onError?: (code: string) => void
}

function readCookie(name: string): string {
  if (typeof document === 'undefined') return ''
  const m = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return m ? decodeURIComponent(m[1]) : ''
}

export function usePitchTextStream() {
  const apiBase = useRuntimeConfig().public.apiBase as string

  let ws: WebSocket | null = null
  let ready = false
  const busy = ref(false)
  const streamingReply = ref('')

  async function connect(simId: string, cb: PitchTextStreamCallbacks): Promise<boolean> {
    const token = readCookie('pf_token')
    if (!token) return false
    const wsUrl = apiBase.replace(/^http/, 'ws')
      + `/api/v1/commercial/pitch-sims/${simId}/stream?token=${encodeURIComponent(token)}`

    ready = false
    streamingReply.value = ''
    busy.value = false

    const opened = await new Promise<boolean>((resolve) => {
      try {
        ws = new WebSocket(wsUrl)
      } catch {
        resolve(false)
        return
      }
      const failTimer = setTimeout(() => resolve(false), 8000)

      ws.onmessage = (ev) => {
        let msg: any
        try { msg = JSON.parse(String(ev.data)) } catch { return }
        switch (msg.type) {
          case 'ready':
            clearTimeout(failTimer)
            ready = true
            cb.onReady()
            resolve(true)
            break
          case 'delta':
            busy.value = true
            streamingReply.value += msg.delta || ''
            cb.onDelta(msg.delta || '')
            break
          case 'turn_complete':
            busy.value = false
            streamingReply.value = msg.reply || streamingReply.value
            cb.onTurnComplete({
              reply: msg.reply || '',
              action: msg.action || 'continue',
              ended: !!msg.ended,
              outcome: msg.outcome || 'in_progress',
              appointmentSlot: msg.appointmentSlot,
              reason: msg.reason,
              interrupted: !!msg.interrupted,
            })
            streamingReply.value = ''
            break
          case 'interrupted':
            busy.value = false
            cb.onInterrupted()
            break
          case 'ended':
            busy.value = false
            cb.onEnded(msg.outcome || 'manual', msg.appointmentSlot)
            break
          case 'error':
            clearTimeout(failTimer)
            busy.value = false
            cb.onError?.(msg.code || 'error')
            if (!ready) resolve(false)
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
        busy.value = false
        cb.onClosed()
      }
    })

    if (!opened) {
      stop()
      return false
    }
    return true
  }

  function sendUser(text: string) {
    if (!text.trim() || ws?.readyState !== WebSocket.OPEN) return
    busy.value = true
    streamingReply.value = ''
    ws.send(JSON.stringify({ type: 'user', text: text.trim() }))
  }

  function interrupt() {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'interrupt' }))
    }
    if (typeof window !== 'undefined') {
      try { window.speechSynthesis?.cancel() } catch { /* ignore */ }
    }
    busy.value = false
  }

  function end() {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'end' }))
    }
    stop()
  }

  function stop() {
    ready = false
    busy.value = false
    streamingReply.value = ''
    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.onmessage = null
      try { ws.close() } catch { /* ignore */ }
      ws = null
    }
  }

  return { connect, sendUser, interrupt, end, stop, busy, streamingReply }
}
