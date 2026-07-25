import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const kind = getRouterParam(event, 'kind')
  return proxyApi(event, `/api/v1/admin/agent-prompts/${kind}/versions`)
})
