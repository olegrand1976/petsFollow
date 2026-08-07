import { clearAuthCookies, revokeIssuedTokens } from '~/server/utils/api'

/** Full-document logout: clear httpOnly cookies then redirect (no Vue required). */
export default defineEventHandler(async (event) => {
  await revokeIssuedTokens(event)
  clearAuthCookies(event)
  return sendRedirect(event, '/login', 302)
})
