import { proxyApi } from '~/server/utils/api'

function appendPaging(params: URLSearchParams, q: Record<string, unknown>) {
  if (typeof q.limit === 'string' && q.limit) params.set('limit', q.limit)
  if (typeof q.offset === 'string' && q.offset) params.set('offset', q.offset)
  if (typeof q.eventType === 'string' && q.eventType) params.set('eventType', q.eventType)
  if (typeof q.commercialId === 'string' && q.commercialId) params.set('commercialId', q.commercialId)
}

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  appendPaging(params, q as Record<string, unknown>)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/commercial/filiation/events${qs ? `?${qs}` : ''}`)
})
