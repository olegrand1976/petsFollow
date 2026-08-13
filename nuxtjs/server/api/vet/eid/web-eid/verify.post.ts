export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const origin = getRequestHeader(event, 'origin') || ''
  return proxyApi(event, '/api/v1/vet/eid/web-eid/verify', {
    method: 'POST',
    body,
    headers: origin ? { 'X-PF-Web-Eid-Origin': origin } : {},
  })
})
