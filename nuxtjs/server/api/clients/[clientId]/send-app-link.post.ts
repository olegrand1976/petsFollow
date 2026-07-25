import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const clientId = getRouterParam(event, 'clientId')
  return proxyApi(event, `/api/v1/clients/${clientId}/send-app-link`, { method: 'POST' })
})
