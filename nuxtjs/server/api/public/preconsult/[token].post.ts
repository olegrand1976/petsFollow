import { proxyPublicApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getRouterParam(event, 'token')
  const body = await readBody(event)
  return proxyPublicApi(event, `/api/v1/public/preconsult/${encodeURIComponent(token || '')}`, {
    method: 'POST',
    body,
  })
})
