import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  if (typeof query.siteId === 'string' && query.siteId) params.set('siteId', query.siteId)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/schedule${qs ? `?${qs}` : ''}`)
})
