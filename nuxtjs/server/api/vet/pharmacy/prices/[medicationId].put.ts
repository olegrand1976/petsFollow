import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const medicationId = getRouterParam(event, 'medicationId')
  const body = await readBody(event)
  return proxyApi(event, `/api/v1/vet/pharmacy/prices/${medicationId}`, { method: 'PUT', body })
})
