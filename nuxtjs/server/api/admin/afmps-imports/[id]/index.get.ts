import { proxyApi } from '../../../../utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const q = getQuery(event)
  const parts = ['limit', 'offset', 'collision']
    .filter(k => q[k] != null && String(q[k]) !== '')
    .map(k => `${k}=${encodeURIComponent(String(q[k]))}`)
  const path = parts.length
    ? `/api/v1/admin/afmps-imports/${id}?${parts.join('&')}`
    : `/api/v1/admin/afmps-imports/${id}`
  return proxyApi(event, path)
})
