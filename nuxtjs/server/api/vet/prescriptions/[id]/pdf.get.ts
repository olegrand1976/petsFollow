import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'id required' })
  }
  return proxyBinary(event, `/api/v1/vet/prescriptions/${id}/pdf`, {
    contentDispositionFallback: 'inline; filename="consignes.pdf"',
  })
})
