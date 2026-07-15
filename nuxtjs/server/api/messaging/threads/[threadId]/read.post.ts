import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const threadId = getRouterParam(event, 'threadId')
  return $fetch(`${apiBase()}/api/v1/messaging/threads/${threadId}/read`, {
    method: 'POST',
    headers: apiHeaders(event),
  })
})
