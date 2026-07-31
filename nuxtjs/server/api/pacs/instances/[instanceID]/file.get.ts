import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const instanceID = getRouterParam(event, 'instanceID')
  if (!instanceID) throw createError({ statusCode: 400, statusMessage: 'instanceID required' })
  return proxyBinary(event, `/api/v1/pacs/instances/${instanceID}/file`)
})
