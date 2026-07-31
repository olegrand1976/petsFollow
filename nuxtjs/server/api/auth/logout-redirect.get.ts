import { clearAuthCookies } from '~/server/utils/api'

/** Full-document logout: clear httpOnly cookies then redirect (no Vue required). */
export default defineEventHandler((event) => {
  clearAuthCookies(event)
  return sendRedirect(event, '/login', 302)
})
