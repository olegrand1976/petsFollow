import { proxyApi } from '../../../utils/api'

export default defineEventHandler(async (event) => {
  return proxyApi(event, '/api/v1/admin/afmps-imports')
})
