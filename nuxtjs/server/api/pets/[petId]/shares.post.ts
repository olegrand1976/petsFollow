import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/pets/${petId}/shares`, { method: 'POST', body })
})
