import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  for (const key of ['depositId', 'medicationId', 'band', 'status', 'q'] as const) {
    const v = query[key]
    if (typeof v === 'string' && v) params.set(key, v)
  }
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/pharmacy/batches${qs ? `?${qs}` : ''}`)
})
