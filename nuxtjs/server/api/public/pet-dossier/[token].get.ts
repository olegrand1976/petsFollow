import { proxyPublicApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getRouterParam(event, 'token')
  return proxyPublicApi(event, `/api/v1/public/pet-dossier/${encodeURIComponent(token || '')}`, { method: 'GET' })
})
