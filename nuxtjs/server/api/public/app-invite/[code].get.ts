import { proxyPublicApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const code = getRouterParam(event, 'code')
  return proxyPublicApi(event, `/api/v1/public/app-invite/${encodeURIComponent(code || '')}`, { method: 'GET' })
})
