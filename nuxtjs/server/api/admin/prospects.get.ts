import { apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  return $fetch(`${config.apiBase}/api/v1/admin/prospects`, {
    headers: apiHeaders(event),
    query: getQuery(event),
  })
})
