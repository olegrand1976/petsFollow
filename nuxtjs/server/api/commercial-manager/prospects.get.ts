import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  for (const [k, v] of Object.entries(q)) {
    if (v != null && v !== '') params.set(k, String(v))
  }
  const qs = params.toString()
  return proxyApi(event, `/api/v1/commercial-manager/prospects${qs ? `?${qs}` : ''}`)
})
