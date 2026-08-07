import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  const documentId = getRouterParam(event, 'documentId')
  if (!petId || !documentId) {
    throw createError({ statusCode: 400, statusMessage: 'petId and documentId required' })
  }
  return proxyBinary(
    event,
    `/api/v1/pets/${petId}/documents/${documentId}/download`,
    { contentDispositionFallback: 'inline; filename="document"' },
  )
})
