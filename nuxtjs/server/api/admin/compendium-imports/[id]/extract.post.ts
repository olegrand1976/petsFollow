import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const q = getQuery(event)
  const restart = q.restart === '1' || q.restart === 'true' || q.restart === true
  const path = restart
    ? `/api/v1/admin/compendium-imports/${id}/extract?restart=1`
    : `/api/v1/admin/compendium-imports/${id}/extract`
  return proxyApi(event, path, { method: 'POST' })
})
