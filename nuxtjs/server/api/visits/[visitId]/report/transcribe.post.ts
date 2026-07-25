import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  const contentType = getHeader(event, 'content-type') || 'multipart/form-data'
  const body = await readRawBody(event, false)
  return proxyUpload(event, `/api/v1/visits/${visitId}/report/transcribe`, body, contentType)
})
