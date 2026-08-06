import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  if (query.includeInactive === '1' || query.includeInactive === 'true') {
    params.set('includeInactive', '1')
  }
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/sites${qs ? `?${qs}` : ''}`)
})
