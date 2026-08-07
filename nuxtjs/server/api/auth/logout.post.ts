import { clearAuthCookies, revokeIssuedTokens } from '~/server/utils/api'

/** Les cookies auth sont httpOnly : seule la BFF peut les supprimer. */
export default defineEventHandler(async (event) => {
  await revokeIssuedTokens(event)
  clearAuthCookies(event)
  return { data: { loggedOut: true } }
})
