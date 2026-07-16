import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  if (!petId) {
    throw createError({ statusCode: 400, statusMessage: 'petId required' })
  }

  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event)
  return proxyUpload(event, `/api/v1/pets/${petId}/photo`, body, contentType)
})
