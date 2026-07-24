import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  const granteeId = getRouterParam(event, 'granteeId')
  return proxyApi(event, `/api/v1/pets/${petId}/shares/${granteeId}`, { method: 'DELETE' })
})
