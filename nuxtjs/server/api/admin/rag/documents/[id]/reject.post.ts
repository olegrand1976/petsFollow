import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const body = await readBody(event).catch(() => ({}))
  return proxyApi(event, `/api/v1/admin/rag/documents/${id}/reject`, { method: 'POST', body })
})
