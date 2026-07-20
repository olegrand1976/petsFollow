import { proxyApi } from '../../utils/proxy'

export default defineEventHandler(async (event) => {
  return proxyApi(event, '/api/v1/admin/commercial-managers')
})
