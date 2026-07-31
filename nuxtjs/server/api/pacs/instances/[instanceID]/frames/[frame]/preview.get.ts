import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const instanceID = getRouterParam(event, 'instanceID')
  const frame = getRouterParam(event, 'frame') || '0'
  if (!instanceID) throw createError({ statusCode: 400, statusMessage: 'instanceID required' })
  return proxyBinary(event, `/api/v1/pacs/instances/${instanceID}/frames/${frame}/preview`)
})
