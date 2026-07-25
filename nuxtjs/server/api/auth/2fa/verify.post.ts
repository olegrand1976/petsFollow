import { proxyPublicApi, absorbAuthTokens } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const res = await proxyPublicApi(event, '/api/v1/auth/2fa/verify', { method: 'POST', body })
  return absorbAuthTokens(event, res)
})
