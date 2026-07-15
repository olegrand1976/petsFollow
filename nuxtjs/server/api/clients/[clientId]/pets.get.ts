import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const clientId = getRouterParam(event, 'clientId')
  return $fetch(`${apiBase()}/api/v1/clients/${clientId}/pets`, { headers: apiHeaders(event) })
})
