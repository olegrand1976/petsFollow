import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const status = typeof query.status === 'string' ? query.status : 'requested'
  return proxyApi(event, `/api/v1/vet/visits?status=${encodeURIComponent(status)}`)
})
