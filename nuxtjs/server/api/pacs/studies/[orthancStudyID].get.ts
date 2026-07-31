import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'orthancStudyID')
  if (!id) throw createError({ statusCode: 400, statusMessage: 'id required' })
  return proxyApi(event, `/api/v1/pacs/studies/${id}`)
})
