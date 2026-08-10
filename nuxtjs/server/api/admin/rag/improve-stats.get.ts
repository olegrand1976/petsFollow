import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const days = q.days ? `?days=${encodeURIComponent(String(q.days))}` : ''
  return proxyApi(event, `/api/v1/admin/rag/improve-stats${days}`)
})
