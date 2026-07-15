import { apiBase, authHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return $fetch(`${apiBase()}/api/v1/vet/profile`, { headers: authHeaders(event) })
})
