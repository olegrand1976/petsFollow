import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'visitId')
  return proxyApi(event, `/api/v1/visits/${id}`, { method: 'DELETE' })
})
