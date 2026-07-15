import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const threadId = getRouterParam(event, 'threadId')
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/messaging/threads/${threadId}/messages`, {
    method: 'POST', headers: apiHeaders(event), body,
  })
})
