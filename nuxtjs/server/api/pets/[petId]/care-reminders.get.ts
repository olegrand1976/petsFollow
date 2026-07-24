import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  return proxyApi(event, `/api/v1/pets/${petId}/care-reminders`)
})
