import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  if (!petId) {
    throw createError({ statusCode: 400, statusMessage: 'petId required' })
  }
  return proxyBinary(event, `/api/v1/pets/${petId}/health-book`, {
    contentDispositionFallback: 'inline; filename="health-book.pdf"',
  })
})
