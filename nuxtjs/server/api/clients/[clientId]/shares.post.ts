import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const clientId = getRouterParam(event, 'clientId')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/clients/${clientId}/shares`, { method: 'POST', body })
})
