import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event, false)
  return proxyUpload(event, '/api/v1/admin/client-imports', body, contentType)
})
