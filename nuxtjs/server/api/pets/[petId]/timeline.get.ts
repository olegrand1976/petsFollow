import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  return $fetch(`${apiBase()}/api/v1/pets/${petId}/timeline`, { headers: apiHeaders(event) })
})
