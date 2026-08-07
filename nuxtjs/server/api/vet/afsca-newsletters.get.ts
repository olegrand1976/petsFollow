import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  if (typeof q.locale === 'string' && q.locale) params.set('locale', q.locale)
  const limitRaw = q.limit
  if (typeof limitRaw === 'string' && limitRaw) params.set('limit', limitRaw)
  else if (typeof limitRaw === 'number' && Number.isFinite(limitRaw)) params.set('limit', String(limitRaw))
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/afsca-newsletters${qs ? `?${qs}` : ''}`)
})
