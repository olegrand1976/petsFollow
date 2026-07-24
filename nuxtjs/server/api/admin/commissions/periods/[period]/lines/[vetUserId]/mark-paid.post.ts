import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const period = getRouterParam(event, 'period')
  const vetUserId = getRouterParam(event, 'vetUserId')
  return proxyApi(event, `/api/v1/admin/commissions/periods/${period}/lines/${vetUserId}/mark-paid`, {
    method: 'POST',
  })
})
