import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  const studyId = getRouterParam(event, 'studyId')
  if (!petId) throw createError({ statusCode: 400, statusMessage: 'petId required' })
  if (!studyId) throw createError({ statusCode: 400, statusMessage: 'studyId required' })
  return proxyApi(event, `/api/v1/pets/${petId}/pacs/studies/${studyId}/comments`)
})
