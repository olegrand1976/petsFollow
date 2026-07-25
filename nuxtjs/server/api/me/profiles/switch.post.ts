import { proxyApi, absorbAuthTokens } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const res = await proxyApi(event, '/api/v1/me/profiles/switch', {
    method: 'POST',
    body: await readBody(event),
  })
  return absorbAuthTokens(event, res)
})
