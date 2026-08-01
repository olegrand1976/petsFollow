import { proxyApi } from '../../../../utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  return proxyApi(event, `/api/v1/admin/afmps-imports/${id}`, { method: 'DELETE' })
})
