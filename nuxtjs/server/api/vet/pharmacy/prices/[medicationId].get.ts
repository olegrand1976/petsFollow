import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const medicationId = getRouterParam(event, 'medicationId')
  if (!medicationId) {
    throw createError({ statusCode: 400, statusMessage: 'medicationId required' })
  }
  return proxyApi(event, `/api/v1/vet/pharmacy/prices/${medicationId}`)
})
