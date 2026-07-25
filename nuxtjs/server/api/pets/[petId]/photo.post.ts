import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  if (!petId) {
    throw createError({ statusCode: 400, statusMessage: 'petId required' })
  }

  const contentType = getHeader(event, 'content-type') || ''
  // false = Buffer brut — l'UTF-8 par défaut corrompt PNG/JPEG (0x89 → U+FFFD)
  const body = await readRawBody(event, false)
  return proxyUpload(event, `/api/v1/pets/${petId}/photo`, body, contentType)
})
