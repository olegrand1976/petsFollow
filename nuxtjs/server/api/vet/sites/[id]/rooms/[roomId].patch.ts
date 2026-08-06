import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const roomId = getRouterParam(event, 'roomId')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/vet/sites/${id}/rooms/${roomId}`, { method: 'PATCH', body })
})
