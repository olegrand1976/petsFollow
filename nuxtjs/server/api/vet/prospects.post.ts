import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return proxyApi(event, '/api/v1/vet/prospects', { method: 'POST', body })
})
