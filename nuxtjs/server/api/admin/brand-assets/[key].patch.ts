import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const key = getRouterParam(event, 'key')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/admin/brand-assets/${encodeURIComponent(key || '')}`, {
    method: 'PATCH',
    body,
  })
})
