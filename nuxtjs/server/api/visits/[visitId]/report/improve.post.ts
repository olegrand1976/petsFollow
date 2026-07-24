import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  return proxyApi(event, `/api/v1/visits/${visitId}/report/improve`, { method: 'POST' })
})
