import { apiBase, authHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const clientId = getRouterParam(event, 'clientId')
  return $fetch(`${apiBase()}/api/v1/clients/${clientId}`, { headers: authHeaders(event) })
})
