import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/vet/notification-preferences`, {
    method: 'PUT',
    headers: apiHeaders(event),
    body,
  })
})
