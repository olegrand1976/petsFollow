import { authHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  return $fetch(`${config.apiBase}/api/v1/auth/2fa/setup`, { method: 'POST', headers: authHeaders(event) })
})
