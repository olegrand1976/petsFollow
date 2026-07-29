import { proxyApi } from '../../utils/api'

/** Consentement CGU/privacy (clients provisionnés / import). */
export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return proxyApi(event, '/api/v1/me/accept-terms', { method: 'POST', body })
})
