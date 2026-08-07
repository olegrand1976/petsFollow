import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getRouterParam(event, 'token')
  return proxyApi(event, `/api/v1/public/proforma/${encodeURIComponent(token || '')}/accept`, {
    method: 'POST',
  })
})
