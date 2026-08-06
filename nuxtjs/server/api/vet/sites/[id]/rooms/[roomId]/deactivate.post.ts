import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const roomId = getRouterParam(event, 'roomId')
  return proxyApi(event, `/api/v1/vet/sites/${id}/rooms/${roomId}/deactivate`, { method: 'POST' })
})
