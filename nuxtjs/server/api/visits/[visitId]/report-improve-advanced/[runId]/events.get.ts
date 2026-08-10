import { createError, getRequestHeader, getRouterParam, sendStream, setResponseHeaders, setResponseStatus } from 'h3'
import { apiBase, apiHeaders } from '../../../../../utils/api'

/**
 * SSE proxy (no body buffering) — EventSource same-origin with httpOnly cookies.
 */
export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  const runId = getRouterParam(event, 'runId')
  if (!visitId || !runId) {
    throw createError({ statusCode: 400, statusMessage: 'missing_ids' })
  }
  const lastEventId = getRequestHeader(event, 'last-event-id')
  const headers: Record<string, string> = {
    ...apiHeaders(event),
    Accept: 'text/event-stream',
  }
  if (lastEventId) headers['Last-Event-ID'] = lastEventId

  const res = await fetch(`${apiBase()}/api/v1/visits/${visitId}/report/improve-advanced/${runId}/events`, {
    method: 'GET',
    headers,
  })
  if (!res.ok || !res.body) {
    const text = await res.text().catch(() => '')
    throw createError({ statusCode: res.status || 502, statusMessage: text || 'sse_upstream' })
  }
  setResponseStatus(event, 200)
  setResponseHeaders(event, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    Connection: 'keep-alive',
    'X-Accel-Buffering': 'no',
  })
  return sendStream(event, res.body)
})
