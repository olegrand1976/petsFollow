import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  if (!visitId) {
    throw createError({ statusCode: 400, statusMessage: 'missing_visit_id' })
  }
  const q = getQuery(event)
  const strip = q.stripCitations === '1' || q.stripCitations === 'true' || q.stripCitations === true
  const path = strip
    ? `/api/v1/visits/${visitId}/report/pdf?stripCitations=1`
    : `/api/v1/visits/${visitId}/report/pdf`
  return proxyBinary(event, path, {
    method: 'GET',
  })
})
