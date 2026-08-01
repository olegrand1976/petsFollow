import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const userId = getRouterParam(event, 'userId')
  return proxyApi(event, `/api/v1/research/groups/${id}/members/${userId}`, { method: 'DELETE' })
})
