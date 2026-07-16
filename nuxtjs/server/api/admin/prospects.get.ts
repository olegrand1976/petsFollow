import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return proxyApi(event, '/api/v1/admin/prospects', {
    query: getQuery(event) as Record<string, unknown>,
  })
})
