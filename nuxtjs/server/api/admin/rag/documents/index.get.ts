import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const status = typeof query.status === 'string' ? query.status : ''
  const suffix = status ? `?status=${encodeURIComponent(status)}` : ''
  return proxyApi(event, `/api/v1/admin/rag/documents${suffix}`)
})
