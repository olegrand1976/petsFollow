import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  if (typeof q.q === 'string' && q.q) params.set('q', q.q)
  if (typeof q.branchId === 'string' && q.branchId) params.set('branchId', q.branchId)
  if (typeof q.commercialId === 'string' && q.commercialId) params.set('commercialId', q.commercialId)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/admin/filiation${qs ? `?${qs}` : ''}`)
})
