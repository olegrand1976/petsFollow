import { proxyPublicApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return proxyPublicApi(event, '/api/v1/auth/confirm-email', { method: 'POST', body })
})
