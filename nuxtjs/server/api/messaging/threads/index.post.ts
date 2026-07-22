import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/messaging/threads`, {
    method: 'POST',
    headers: apiHeaders(event),
    body,
  })
})
