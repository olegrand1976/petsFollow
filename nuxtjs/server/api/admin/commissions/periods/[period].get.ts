import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const period = getRouterParam(event, 'period')
  return proxyApi(event, `/api/v1/admin/commissions/periods/${period}`)
})
