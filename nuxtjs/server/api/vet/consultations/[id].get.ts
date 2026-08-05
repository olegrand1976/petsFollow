import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  if (!id?.trim()) {
    throw createError({ statusCode: 404, statusMessage: 'Not Found' })
  }
  return proxyApi(event, `/api/v1/vet/consultations/${encodeURIComponent(id.trim())}`)
})
