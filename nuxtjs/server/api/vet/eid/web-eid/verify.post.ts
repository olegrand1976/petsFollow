import { proxyApi, webEidUpstreamHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  return proxyApi(event, '/api/v1/vet/eid/web-eid/verify', {
    method: 'POST',
    body,
    headers: webEidUpstreamHeaders(event),
  })
})
