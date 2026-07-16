import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const threadId = getRouterParam(event, 'threadId')
  if (!threadId) {
    throw createError({ statusCode: 400, statusMessage: 'threadId required' })
  }
  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event)
  return proxyUpload(event, `/api/v1/messaging/threads/${threadId}/messages/media`, body, contentType)
})
