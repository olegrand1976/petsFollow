import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/vet/pharmacy/daf/${id}/cancel`, { method: 'POST', body: body ?? {} })
})
