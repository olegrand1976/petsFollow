import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const documentId = getRouterParam(event, 'documentId')
  if (!documentId) {
    throw createError({ statusCode: 400, statusMessage: 'documentId required' })
  }
  return $fetch(`${apiBase()}/api/v1/pets/documents/${documentId}`, {
    method: 'DELETE',
    headers: apiHeaders(event),
  })
})
