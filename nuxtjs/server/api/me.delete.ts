import { proxyApi, clearAuthCookies } from '~/server/utils/api'

/** Suppression de compte Pro (RGPD) : effacement API puis purge de la session. */
export default defineEventHandler(async (event) => {
  const res = await proxyApi(event, '/api/v1/me', { method: 'DELETE' })
  clearAuthCookies(event)
  return res
})
