import { apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')
  return $fetch(`${config.apiBase}/api/v1/commercial/prospects/${id}`, {
    method: 'DELETE',
    headers: apiHeaders(event),
  })
})
