import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return proxyApi(event, '/api/v1/messaging/threads', {
    method: 'POST',
    body: await readBody(event),
  })
})
