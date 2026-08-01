import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  const panelId = getRouterParam(event, 'panelId')
  return proxyApi(event, `/api/v1/pets/${petId}/lab-panels/${panelId}`, {
    method: 'DELETE',
  })
})
