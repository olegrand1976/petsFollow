import { apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const config = useRuntimeConfig()
  return $fetch(`${config.apiBase}/api/v1/auth/2fa/confirm`, { method: 'POST', body, headers: apiHeaders(event) })
})
