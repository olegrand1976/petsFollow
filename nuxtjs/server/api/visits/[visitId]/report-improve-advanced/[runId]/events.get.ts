import { proxySSE } from '~/server/utils/api'

/**
 * SSE proxy (no body buffering) — EventSource same-origin with httpOnly cookies.
 * Refresh+retry lives in proxySSE (bffUpstreamCalls invariant).
 */
export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  const runId = getRouterParam(event, 'runId')
  if (!visitId || !runId) {
    throw createError({ statusCode: 400, statusMessage: 'missing_ids' })
  }
  return proxySSE(event, `/api/v1/visits/${visitId}/report/improve-advanced/${runId}/events`)
})
