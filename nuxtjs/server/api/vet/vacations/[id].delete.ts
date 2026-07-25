import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  return proxyApi(event, `/api/v1/vet/vacations/${id}`, { method: 'DELETE' })
})
