import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  return proxyBinary(event, `/api/v1/vet/pharmacy/inventory/sessions/${id}/export.csv`, {
    contentDispositionFallback: 'attachment; filename="inventory.csv"',
  })
})
