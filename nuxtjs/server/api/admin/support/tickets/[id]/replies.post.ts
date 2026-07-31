import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const body = await readBody(event).catch(() => ({}))
  const data = await proxyApi(event, `/api/v1/admin/support/tickets/${id}/replies`, { method: 'POST', body })
  setResponseStatus(event, 201)
  return data
})
