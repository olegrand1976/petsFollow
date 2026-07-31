import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const practiceId = getRouterParam(event, 'practiceId')
  return proxyApi(event, `/api/v1/admin/invoicing/connections/${encodeURIComponent(practiceId || '')}/saas-draft`, {
    method: 'POST',
  })
})
