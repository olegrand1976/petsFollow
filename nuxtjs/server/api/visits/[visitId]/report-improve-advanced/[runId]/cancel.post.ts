import { proxyApi } from '../../../../../utils/api'

export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  const runId = getRouterParam(event, 'runId')
  return proxyApi(event, `/api/v1/visits/${visitId}/report/improve-advanced/${runId}/cancel`, {
    method: 'POST',
  })
})
