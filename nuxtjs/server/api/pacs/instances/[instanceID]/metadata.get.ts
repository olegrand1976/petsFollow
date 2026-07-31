import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const instanceID = getRouterParam(event, 'instanceID')
  if (!instanceID) throw createError({ statusCode: 400, statusMessage: 'instanceID required' })
  return proxyApi(event, `/api/v1/pacs/instances/${instanceID}/metadata`)
})
