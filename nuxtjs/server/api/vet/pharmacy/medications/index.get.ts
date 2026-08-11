import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const params = new URLSearchParams()
  if (typeof query.letter === 'string') params.set('letter', query.letter)
  if (typeof query.limit === 'string') params.set('limit', query.limit)
  if (typeof query.offset === 'string') params.set('offset', query.offset)
  const qs = params.toString()
  return proxyApi(event, `/api/v1/vet/pharmacy/medications${qs ? `?${qs}` : ''}`)
})
