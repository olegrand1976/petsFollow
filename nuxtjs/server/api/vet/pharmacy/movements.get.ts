import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  if (typeof query.limit === 'string') params.set('limit', query.limit)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/pharmacy/movements${qs ? `?${qs}` : ''}`)
})
