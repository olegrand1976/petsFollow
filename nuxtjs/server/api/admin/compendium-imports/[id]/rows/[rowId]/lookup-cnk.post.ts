import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const rowId = getRouterParam(event, 'rowId')
  const body = await readBody(event).catch(() => ({}))
  return proxyApi(event, `/api/v1/admin/compendium-imports/${id}/rows/${rowId}/lookup-cnk`, {
    method: 'POST',
    body: body ?? {},
  })
})
