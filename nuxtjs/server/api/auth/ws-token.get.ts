import { refreshAccessToken } from '~/server/utils/api'

/**
 * Fournit le token d'accès courant (TTL court) pour les WebSockets pitch,
 * les cookies auth étant httpOnly et illisibles côté client.
 */
export default defineEventHandler(async (event) => {
  let token = getCookie(event, 'pf_token')
  if (!token) {
    const pair = await refreshAccessToken(event)
    token = pair?.accessToken
  }
  if (!token) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }
  return { data: { token } }
})
