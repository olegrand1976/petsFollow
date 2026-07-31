import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  const body = await readBody(event).catch(() => undefined)
  return proxyApi(event, `/api/v1/visits/${visitId}/report/improve`, {
    method: 'POST',
    body: body ?? {},
  })
})
