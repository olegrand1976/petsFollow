import type { AgentStep } from '~/components/pro/ProAgentLoader.vue'

export type AdvancedImproveHandlers = {
  onStep?: (step: AgentStep) => void
  onFinal?: (report: string) => void
  onError?: (code: string) => void
  onDone?: () => void
}

/**
 * Starts an advanced improve run and consumes SSE via EventSource (cookies).
 * cancelRemote() POSTs cancel then settles the in-flight promise (unmount / Annuler).
 */
export function useAdvancedImproveStream() {
  let es: EventSource | null = null
  let activeVisitId: string | null = null
  let activeRunId: string | null = null
  let intentionalClose = false
  let finish: ((code?: string) => void) | null = null

  function stop() {
    es?.close()
    es = null
  }

  function clearActive() {
    activeVisitId = null
    activeRunId = null
    finish = null
  }

  async function cancelRemote() {
    const visitId = activeVisitId
    const runId = activeRunId
    intentionalClose = true
    if (visitId && runId) {
      try {
        await $fetch(`/api/visits/${visitId}/report-improve-advanced/${runId}/cancel`, {
          method: 'POST',
        })
      } catch { /* best-effort */ }
    }
    finish?.('cancelled')
    stop()
    clearActive()
  }

  async function start(
    visitId: string,
    body: { sourceText?: string; targetLocale?: string },
    handlers: AdvancedImproveHandlers = {},
  ) {
    if (activeRunId) {
      await cancelRemote()
    } else {
      stop()
    }
    intentionalClose = false

    const res = await $fetch<{ data?: { runId?: string }; runId?: string }>(
      `/api/visits/${visitId}/report-improve-advanced`,
      { method: 'POST', body },
    )
    const runId = res?.data?.runId || res?.runId
    if (!runId) throw new Error('run_id_missing')

    activeVisitId = visitId
    activeRunId = runId

    await new Promise<void>((resolve) => {
      let settled = false
      finish = (code?: string) => {
        if (settled) return
        settled = true
        if (code) handlers.onError?.(code)
        else handlers.onDone?.()
        stop()
        clearActive()
        resolve()
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
        if (ev instanceof MessageEvent && typeof ev.data === 'string' && ev.data.length > 0) {
          try {
            const data = JSON.parse(ev.data || '{}') as { errorCode?: string }
            finish?.(data.errorCode || 'error')
          } catch {
            finish?.('error')
          }
          return
        }
        if (intentionalClose || settled) return
        if (es && es.readyState === EventSource.CLOSED) {
          finish?.('sse_closed')
        }
      })
      es.addEventListener('ping', (ev) => {
        try {
          const data = JSON.parse((ev as MessageEvent).data || '{}') as { done?: boolean }
          if (data.done) finish?.()
        } catch { /* ignore */ }
      })
    })
    return runId
  }

  return { start, stop, cancelRemote, get activeRunId() { return activeRunId } }
}
