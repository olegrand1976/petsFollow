import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const practiceID = getRouterParam(event, 'practiceID')
  return proxyApi(event, `/api/v1/commercial/ai-modules/${practiceID}/activate`, { method: 'POST' })
})
