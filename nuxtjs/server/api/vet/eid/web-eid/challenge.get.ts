export default defineEventHandler(async (event) => {
  const origin = getRequestHeader(event, 'origin') || ''
  return proxyApi(event, '/api/v1/vet/eid/web-eid/challenge', {
    method: 'GET',
    headers: origin ? { 'X-PF-Web-Eid-Origin': origin } : {},
  })
})
