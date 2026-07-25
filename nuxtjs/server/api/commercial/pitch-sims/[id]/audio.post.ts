import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const contentType = getHeader(event, 'content-type') || 'multipart/form-data'
  const body = await readRawBody(event, false)
  return proxyUpload(event, `/api/v1/commercial/pitch-sims/${id}/audio`, body, contentType)
})
