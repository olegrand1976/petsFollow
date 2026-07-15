import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return $fetch(`${apiBase()}/api/v1/vet/notification-preferences`, { headers: apiHeaders(event) })
})
