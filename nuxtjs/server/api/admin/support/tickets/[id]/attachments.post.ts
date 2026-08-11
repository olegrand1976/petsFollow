import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const contentType = getHeader(event, 'content-type') || 'multipart/form-data'
  const body = await readRawBody(event, false)
  const data = await proxyUpload(event, `/api/v1/admin/support/tickets/${id}/attachments`, body, contentType)
  setResponseStatus(event, 201)
  return data
})
