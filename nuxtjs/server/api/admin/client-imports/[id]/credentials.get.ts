import { apiBase, apiHeaders, proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const query = getQuery(event)
  const token = typeof query.token === 'string' ? query.token : ''
  if (!token) {
    throw createError({ statusCode: 400, statusMessage: 'token_required' })
  }

  // Touch a JSON route first so proxyApi can refresh an expired access token.
  await proxyApi(event, `/api/v1/admin/client-imports/${id}`).catch(() => null)

  const url = `${apiBase()}/api/v1/admin/client-imports/${id}/credentials?token=${encodeURIComponent(token)}`
  const res = await fetch(url, { headers: apiHeaders(event) })
  if (!res.ok) {
    const text = await res.text()
    throw createError({ statusCode: res.status, statusMessage: text || 'credentials_unavailable' })
  }

  const buf = Buffer.from(await res.arrayBuffer())
  setHeader(event, 'Content-Type', res.headers.get('Content-Type') || 'text/csv; charset=utf-8')
  setHeader(event, 'Content-Disposition', res.headers.get('Content-Disposition') || 'attachment; filename="client-import-credentials.csv"')
  return buf
})
