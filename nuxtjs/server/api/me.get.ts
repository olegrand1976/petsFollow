import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getCookie(event, 'pf_token')
  const refresh = getCookie(event, 'pf_refresh')
  if (!token && !refresh) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }
  return proxyApi(event, '/api/v1/me')
})
