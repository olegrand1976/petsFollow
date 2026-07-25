import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return proxyApi(event, '/api/v1/me/locale', {
    method: 'PATCH',
    body: await readBody(event),
  })
})
