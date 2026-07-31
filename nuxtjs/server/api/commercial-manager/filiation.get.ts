import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  if (typeof q.q === 'string' && q.q) params.set('q', q.q)
  if (typeof q.commercialId === 'string' && q.commercialId) params.set('commercialId', q.commercialId)
  if (typeof q.limit === 'string' && q.limit) params.set('limit', q.limit)
  if (typeof q.offset === 'string' && q.offset) params.set('offset', q.offset)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/commercial-manager/filiation${qs ? `?${qs}` : ''}`)
})
