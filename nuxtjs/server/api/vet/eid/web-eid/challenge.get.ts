import { proxyApi, webEidUpstreamHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  return proxyApi(event, '/api/v1/vet/eid/web-eid/challenge', {
    method: 'GET',
    headers: webEidUpstreamHeaders(event),
  })
})
