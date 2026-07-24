import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/visits/${visitId}/report`, { method: 'PUT', body })
})
