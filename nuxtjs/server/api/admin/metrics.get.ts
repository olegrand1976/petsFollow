import { apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const q = getQuery(event)
  const config = useRuntimeConfig()
  return $fetch(`${config.apiBase}/api/v1/admin/metrics/overview`, {
    headers: apiHeaders(event),
    query: q,
  })
})
