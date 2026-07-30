import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  return proxyApi(event, '/api/v1/vet/pharmacy/daf/for-visit', { query })
})
