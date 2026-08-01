import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petID = getRouterParam(event, 'petID')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/vet/pets/${petID}/lifecycle`, { method: 'PATCH', body })
})
