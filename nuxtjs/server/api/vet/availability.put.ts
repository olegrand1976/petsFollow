import { apiBase, authHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  if (event.method === 'GET') {
    return $fetch(`${apiBase()}/api/v1/vet/availability`, { headers: authHeaders(event) })
  }
  const body = await readBody(event)
  return $fetch(`${apiBase()}/api/v1/vet/availability`, {
    method: 'PUT', headers: authHeaders(event), body,
  })
})
