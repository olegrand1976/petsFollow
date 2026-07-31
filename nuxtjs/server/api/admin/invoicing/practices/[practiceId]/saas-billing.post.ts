import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const practiceId = getRouterParam(event, 'practiceId')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/admin/invoicing/practices/${encodeURIComponent(practiceId || '')}/saas-billing`, {
    method: 'POST',
    body,
  })
})
