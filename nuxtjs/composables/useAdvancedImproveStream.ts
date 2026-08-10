import type { AgentStep } from '~/components/pro/ProAgentLoader.vue'

export type AdvancedImproveHandlers = {
  onStep?: (step: AgentStep) => void
  onFinal?: (report: string) => void
  onError?: (code: string) => void
  onDone?: () => void
}

/**
 * Starts an advanced improve run and consumes SSE via EventSource (cookies).
 */
export function useAdvancedImproveStream() {
  let es: EventSource | null = null

  function stop() {
    es?.close()
    es = null
  }

  async function start(
    visitId: string,
    body: { sourceText?: string; targetLocale?: string },
    handlers: AdvancedImproveHandlers = {},
  ) {
    stop()
    const res = await $fetch<{ data?: { runId?: string }; runId?: string }>(
      `/api/visits/${visitId}/report-improve-advanced`,
      { method: 'POST', body },
    )
    const runId = res?.data?.runId || res?.runId
    if (!runId) throw new Error('run_id_missing')

    await new Promise<void>((resolve, reject) => {
      let settled = false
      const settleOk = () => {
        if (settled) return
        settled = true
        handlers.onDone?.()
        stop()
        resolve()
      }
      const settleErr = (code: string, asReject = false) => {
        if (settled) return
        settled = true
        handlers.onError?.(code)
        stop()
        if (asReject) reject(new Error(code))
        else resolve()
      }

      es = new EventSource(`/api/visits/${visitId}/report-improve-advanced/${runId}/events`)
      es.addEventListener('step', (ev) => {
        try {
          const data = JSON.parse((ev as MessageEvent).data || '{}') as AgentStep
          handlers.onStep?.(data)
        } catch { /* ignore */ }
      })
      es.addEventListener('final', (ev) => {
        try {
          const data = JSON.parse((ev as MessageEvent).data || '{}') as { report?: string }
          handlers.onFinal?.(data.report || '')
        } catch { /* ignore */ }
      })
      es.addEventListener('error', (ev) => {
        // Named SSE "error" events are MessageEvent with data; native network errors are not.
        if (ev instanceof MessageEvent && typeof ev.data === 'string' && ev.data.length > 0) {
          try {
            const data = JSON.parse(ev.data || '{}') as { errorCode?: string }
            settleErr(data.errorCode || 'error')
          } catch {
            settleErr('error')
          }
          return
        }
        // After stop()/success, closing EventSource fires a native error — ignore if settled.
        if (settled) return
        if (es && es.readyState === EventSource.CLOSED) {
          settleErr('sse_closed', true)
        }
      })
      es.addEventListener('ping', (ev) => {
        try {
          const data = JSON.parse((ev as MessageEvent).data || '{}') as { done?: boolean }
          if (data.done) settleOk()
        } catch { /* ignore */ }
      })
    })
    return runId
  }

  return { start, stop }
}
