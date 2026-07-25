import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return proxyApi(event, '/api/v1/vet/notification-preferences', {
    method: 'PUT',
    body: await readBody(event),
  })
})
