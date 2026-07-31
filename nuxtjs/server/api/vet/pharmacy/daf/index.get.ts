import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  if (typeof query.status === 'string') params.set('status', query.status)
  if (typeof query.year === 'string') params.set('year', query.year)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/pharmacy/daf${qs ? `?${qs}` : ''}`)
})
