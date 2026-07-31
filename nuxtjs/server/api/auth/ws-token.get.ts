import { refreshAccessToken } from '~/server/utils/api'

/**
 * Fournit le token d'accès courant (TTL court) pour les WebSockets pitch,
 * les cookies auth étant httpOnly et illisibles côté client.
 */
export default defineEventHandler(async (event) => {
  let token = getCookie(event, 'pf_token')
  if (!token) {
    const outcome = await refreshAccessToken(event)
    switch (outcome.kind) {
      case 'ok':
        token = outcome.pair.accessToken
        break
      case 'transient':
        throw createError({
          statusCode: 503,
          statusMessage: 'Auth refresh temporarily unavailable',
          data: { reason: 'refresh_unavailable' },
        })
      case 'missing':
      case 'rejected':
        throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
      default: {
        const _exhaustive: never = outcome
        throw _exhaustive
      }
    }
  }
  if (!token) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }
  return { data: { token } }
})
