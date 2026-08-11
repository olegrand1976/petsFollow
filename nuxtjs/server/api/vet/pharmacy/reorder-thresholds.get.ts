import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const medicationId = String(q.medicationId || '').trim()
  const qs = medicationId ? `?medicationId=${encodeURIComponent(medicationId)}` : ''
  return proxyApi(event, `/api/v1/vet/pharmacy/reorder-thresholds${qs}`)
})
