import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  const analyteCode = getRouterParam(event, 'analyteCode')
  return proxyApi(event, `/api/v1/pets/${petId}/lab-analytes/${analyteCode}/trend`)
})
