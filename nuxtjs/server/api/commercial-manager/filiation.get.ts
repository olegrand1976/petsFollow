import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  if (typeof q.q === 'string' && q.q) params.set('q', q.q)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/commercial-manager/filiation${qs ? `?${qs}` : ''}`)
})
