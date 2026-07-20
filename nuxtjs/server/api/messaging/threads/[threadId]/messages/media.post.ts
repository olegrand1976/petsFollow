import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const threadId = getRouterParam(event, 'threadId')
  if (!threadId) {
    throw createError({ statusCode: 400, statusMessage: 'threadId required' })
  }
  const contentType = getHeader(event, 'content-type') || ''
  // false = Buffer brut — l'UTF-8 par défaut corrompt PNG/JPEG (0x89 → U+FFFD)
  const body = await readRawBody(event, false)
  return proxyUpload(event, `/api/v1/messaging/threads/${threadId}/messages/media`, body, contentType)
})
