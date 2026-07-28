import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const qs = new URLSearchParams()
  if (q.status) qs.set('status', String(q.status))
  if (q.petId) qs.set('petId', String(q.petId))
  const suffix = qs.toString() ? `?${qs}` : ''
  return proxyApi(event, `/api/v1/vet/prescriptions${suffix}`)
})
