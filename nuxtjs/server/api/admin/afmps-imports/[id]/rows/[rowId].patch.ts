import { proxyApi } from '../../../../../utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const rowId = getRouterParam(event, 'rowId')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/admin/afmps-imports/${id}/rows/${rowId}`, { method: 'PATCH', body })
})
