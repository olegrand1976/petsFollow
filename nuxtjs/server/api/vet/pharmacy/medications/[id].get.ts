import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'missing_id' })
  }
  return proxyApi(event, `/api/v1/vet/pharmacy/medications/${id}`)
})
