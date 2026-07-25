import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  if (event.method === 'GET') {
    return proxyApi(event, '/api/v1/vet/availability')
  }
  return proxyApi(event, '/api/v1/vet/availability', {
    method: 'PUT',
    body: await readBody(event),
  })
})
