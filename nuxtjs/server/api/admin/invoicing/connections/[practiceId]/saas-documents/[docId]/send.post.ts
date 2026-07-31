import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const practiceId = getRouterParam(event, 'practiceId')
  const docId = getRouterParam(event, 'docId')
  return proxyApi(
    event,
    `/api/v1/admin/invoicing/connections/${encodeURIComponent(practiceId || '')}/saas-documents/${encodeURIComponent(docId || '')}/send`,
    { method: 'POST' },
  )
})
