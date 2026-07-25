import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const practiceID = getRouterParam(event, 'practiceID')
  const body = await readBody(event).catch(() => ({}))
  return proxyApi(event, `/api/v1/admin/ai-modules/${practiceID}`, {
    method: 'PATCH',
    body,
  })
})
