import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const token = getRouterParam(event, 'token')
  if (!id || !token) {
    throw createError({ statusCode: 400, statusMessage: 'token_required' })
  }
  return proxyBinary(
    event,
    `/api/v1/admin/client-imports/${id}/credentials/${encodeURIComponent(token)}`,
    { contentDispositionFallback: 'attachment; filename="client-import-credentials.csv"' },
  )
})
