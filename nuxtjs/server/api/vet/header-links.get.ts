import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const params = new URLSearchParams()
  if (typeof q.locale === 'string' && q.locale) params.set('locale', q.locale)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/header-links${qs ? `?${qs}` : ''}`)
})
