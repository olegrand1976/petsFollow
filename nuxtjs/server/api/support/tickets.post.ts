import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event).catch(() => ({}))
  const data = await proxyApi(event, '/api/v1/support/tickets', { method: 'POST', body })
  setResponseStatus(event, 201)
  return data
})
