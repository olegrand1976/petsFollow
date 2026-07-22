import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  if (!petId) {
    throw createError({ statusCode: 400, statusMessage: 'petId required' })
  }
  return $fetch(`${apiBase()}/api/v1/pets/${petId}/documents`, { headers: apiHeaders(event) })
})
