import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const query = getQuery(event)
  const complete = query.complete === 'true' ? '?complete=true' : ''
  return proxyApi(event, `/api/v1/vet/profile${complete}`, { method: 'PUT', body })
})
