import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  if (typeof query.q === 'string') params.set('q', query.q)
  if (typeof query.limit === 'string') params.set('limit', query.limit)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/pharmacy/medications/search${qs ? `?${qs}` : ''}`)
})
