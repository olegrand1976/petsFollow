import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  for (const key of ['kind', 'includeDeprecated'] as const) {
    const v = query[key]
    if (typeof v === 'string' && v) params.set(key, v)
  }
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/pharmacy/vamreg-refs${qs ? `?${qs}` : ''}`)
})
