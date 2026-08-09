import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  const visitId = String(q.visitId || '').trim()
  const dafId = String(q.dafId || '').trim()
  if (visitId) params.set('visitId', visitId)
  if (dafId) params.set('dafId', dafId)
  const suffix = params.toString() ? `?${params.toString()}` : ''
  return proxyApi(event, `/api/v1/practices/me/invoicing/prefill${suffix}`)
})
