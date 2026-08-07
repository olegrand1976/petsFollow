import { apiBase, apiHeaders, proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const petId = getRouterParam(event, 'petId')
  const documentId = getRouterParam(event, 'documentId')
  if (!petId || !documentId) {
    throw createError({ statusCode: 400, statusMessage: 'petId and documentId required' })
  }

  // Touch JSON route first so proxyApi can refresh an expired access token.
  await proxyApi(event, `/api/v1/pets/${petId}`).catch(() => null)

  const url = `${apiBase()}/api/v1/pets/${encodeURIComponent(petId)}/documents/${encodeURIComponent(documentId)}/download`
  const res = await fetch(url, { headers: apiHeaders(event) })
  if (!res.ok) {
    const text = await res.text()
    throw createError({ statusCode: res.status, statusMessage: text || 'document_unavailable' })
  }

  const buf = Buffer.from(await res.arrayBuffer())
  setHeader(event, 'Content-Type', res.headers.get('Content-Type') || 'application/octet-stream')
  setHeader(event, 'Content-Disposition', res.headers.get('Content-Disposition') || 'inline; filename="document"')
  setHeader(event, 'Cache-Control', 'private, no-store')
  return buf
})
