import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  for (const key of ['limit', 'source', 'importance'] as const) {
    const v = q[key]
    if (typeof v === 'string' && v) params.set(key, v)
    else if (typeof v === 'number' && Number.isFinite(v)) params.set(key, String(v))
  }
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/news${qs ? `?${qs}` : ''}`)
})
