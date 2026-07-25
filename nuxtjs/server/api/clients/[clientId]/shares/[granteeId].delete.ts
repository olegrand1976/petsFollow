import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const clientId = getRouterParam(event, 'clientId')
  const granteeId = getRouterParam(event, 'granteeId')
  return proxyApi(event, `/api/v1/clients/${clientId}/shares/${granteeId}`, { method: 'DELETE' })
})
