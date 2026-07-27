import { apiBase, apiHeaders, proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  if (!petId) {
    throw createError({ statusCode: 400, statusMessage: 'petId required' })
  }

  // Touch JSON route first so proxyApi can refresh an expired access token.
  await proxyApi(event, `/api/v1/pets/${petId}`).catch(() => null)

  const url = `${apiBase()}/api/v1/pets/${petId}/health-book`
  const res = await fetch(url, { headers: apiHeaders(event) })
  if (!res.ok) {
    const text = await res.text()
    throw createError({ statusCode: res.status, statusMessage: text || 'health_book_unavailable' })
  }

  const buf = Buffer.from(await res.arrayBuffer())
  setHeader(event, 'Content-Type', res.headers.get('Content-Type') || 'application/pdf')
  setHeader(event, 'Content-Disposition', res.headers.get('Content-Disposition') || 'inline; filename="health-book.pdf"')
  setHeader(event, 'Cache-Control', 'private, no-store')
  return buf
})
