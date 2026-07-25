import { proxyPublicApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return proxyPublicApi(event, '/api/v1/public/brand-assets', { method: 'GET' })
})
