import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  for (const key of ['species', 'signal', 'country', 'postalCode', 'from', 'to', 'limit'] as const) {
    const v = q[key]
    if (typeof v === 'string' && v) params.set(key, v)
  }
  const qs = params.toString()
  return proxyApi(event, `/api/v1/research/dataroom/events${qs ? `?${qs}` : ''}`)
})
