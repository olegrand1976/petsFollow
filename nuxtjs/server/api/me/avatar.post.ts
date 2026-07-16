import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event)
  return proxyUpload(event, '/api/v1/me/avatar', body, contentType)
})
