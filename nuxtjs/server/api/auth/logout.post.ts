import { clearAuthCookies } from '~/server/utils/api'

/** Les cookies auth sont httpOnly : seule la BFF peut les supprimer. */
export default defineEventHandler((event) => {
  clearAuthCookies(event)
  return { data: { loggedOut: true } }
})
