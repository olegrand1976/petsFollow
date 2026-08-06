import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const query = getQuery(event)
  return proxyApi(event, `/api/v1/commercial/prospects/${id}/activities`, { query })
})
