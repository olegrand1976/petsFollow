import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const res = await proxyApi(event, '/api/v1/vet/pharmacy/delivery-notes', { method: 'POST', body })
  setResponseStatus(event, 201)
  return res
})
