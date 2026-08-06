import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  return proxyApi(event, `/api/v1/commercial/email-templates/${id}`, {
    method: 'PATCH',
    body: await readBody(event),
  })
})
