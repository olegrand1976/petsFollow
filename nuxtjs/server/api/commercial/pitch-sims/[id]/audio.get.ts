import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'missing_sim_id' })
  }
  return proxyBinary(event, `/api/v1/commercial/pitch-sims/${id}/audio`)
})
