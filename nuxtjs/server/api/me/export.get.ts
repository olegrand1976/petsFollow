import { proxyApi } from '~/server/utils/api'

/** Portabilité RGPD : export JSON des données du compte. */
export default defineEventHandler(async (event) => {
  const res = await proxyApi(event, '/api/v1/me/export', { method: 'GET' })
  setHeader(event, 'Content-Disposition', 'attachment; filename="petsfollow-export.json"')
  return res
})
