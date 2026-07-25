import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  if (typeof query.from === 'string') params.set('from', query.from)
  if (typeof query.to === 'string') params.set('to', query.to)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/calendar${qs ? `?${qs}` : ''}`)
})
