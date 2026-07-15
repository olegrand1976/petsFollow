import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return $fetch(`${apiBase()}/api/v1/messaging/threads/read-all`, {
    method: 'POST',
    headers: apiHeaders(event),
  })
})
