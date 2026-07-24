import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event, false)
  const query = getQuery(event)
  const kind = typeof query.kind === 'string' ? query.kind : ''
  const qs = kind ? `?kind=${encodeURIComponent(kind)}` : ''
  return proxyUpload(event, `/api/v1/admin/stripe-catalog/import${qs}`, body, contentType)
})
