import { proxyApi } from '../../../../utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/admin/afmps-imports/${id}/commit`, { method: 'POST', body })
})
