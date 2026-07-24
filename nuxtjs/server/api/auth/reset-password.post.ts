import { proxyPublicApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return proxyPublicApi(event, '/api/v1/auth/reset-password', { method: 'POST', body })
})
