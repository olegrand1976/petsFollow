import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const qs = new URLSearchParams()
  if (q.status) qs.set('status', String(q.status))
  if (q.petId) qs.set('petId', String(q.petId))
  if (q.q) qs.set('q', String(q.q))
  if (q.from) qs.set('from', String(q.from))
  if (q.to) qs.set('to', String(q.to))
  const suffix = qs.toString() ? `?${qs}` : ''
  return proxyApi(event, `/api/v1/vet/prescriptions${suffix}`)
})
