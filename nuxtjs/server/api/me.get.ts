import { authHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getCookie(event, 'pf_token')
  if (!token) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  const config = useRuntimeConfig()
  try {
    return await $fetch(`${config.apiBase}/api/v1/me`, { headers: authHeaders(event) })
  } catch {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }
})
