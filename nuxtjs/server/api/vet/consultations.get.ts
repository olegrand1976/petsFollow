import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  for (const key of ['status', 'q', 'from', 'to', 'hasAudio', 'limit', 'offset'] as const) {
    const v = query[key]
    if (typeof v === 'string' && v.trim()) {
      params.set(key, v.trim())
    }
  }
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/consultations${qs ? `?${qs}` : ''}`)
})
