import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  if (!visitId) {
    throw createError({ statusCode: 400, statusMessage: 'missing_visit_id' })
  }
  return proxyBinary(event, `/api/v1/visits/${visitId}/report/audio`)
})
