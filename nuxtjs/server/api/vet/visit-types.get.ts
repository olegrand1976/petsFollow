import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const qs = new URLSearchParams()
  if (query.active != null) qs.set('active', String(query.active))
  const suffix = qs.toString() ? `?${qs.toString()}` : ''
  return proxyApi(event, `/api/v1/vet/visit-types${suffix}`)
})
