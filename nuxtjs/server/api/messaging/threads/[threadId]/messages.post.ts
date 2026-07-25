import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const threadId = getRouterParam(event, 'threadId')
  return proxyApi(event, `/api/v1/messaging/threads/${threadId}/messages`, {
    method: 'POST',
    body: await readBody(event),
  })
})
