import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event).catch(() => ({}))
  return proxyApi(event, '/api/v1/me/ai-module/feedback', { method: 'POST', body })
})
