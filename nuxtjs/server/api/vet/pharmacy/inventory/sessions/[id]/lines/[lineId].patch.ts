import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const lineId = getRouterParam(event, 'lineId')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/vet/pharmacy/inventory/sessions/${id}/lines/${lineId}`, {
    method: 'PATCH',
    body,
  })
})
