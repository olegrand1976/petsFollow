import { proxyUpload } from '../../../utils/api'

export default defineEventHandler(async (event) => {
  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event)
  return proxyUpload(event, '/api/v1/admin/afmps-imports', body, contentType)
})
