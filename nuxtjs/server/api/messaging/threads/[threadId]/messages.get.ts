import { apiBase, authHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const threadId = getRouterParam(event, 'threadId')
  return $fetch(`${apiBase()}/api/v1/messaging/threads/${threadId}/messages`, { headers: authHeaders(event) })
})
