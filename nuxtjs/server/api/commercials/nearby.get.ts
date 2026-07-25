import { proxyPublicApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  for (const key of ['lat', 'lng', 'postalCode', 'limit', 'radiusKm'] as const) {
    const v = q[key]
    if (v !== undefined && v !== null && String(v).trim() !== '') {
      params.set(key, String(v))
    }
  }
  const qs = params.toString()
  return proxyPublicApi(event, `/api/v1/commercials/nearby${qs ? `?${qs}` : ''}`, { method: 'GET' })
})
