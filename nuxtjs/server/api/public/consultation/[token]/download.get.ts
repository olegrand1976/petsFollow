import { apiBase, localeHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getRouterParam(event, 'token')
  const url = `${apiBase()}/api/v1/public/consultation/${encodeURIComponent(token || '')}/download`
  try {
    const res = await $fetch.raw(url, {
      method: 'GET',
      responseType: 'stream',
      headers: localeHeaders(event),
    })
    const ct = res.headers.get('content-type') || 'application/pdf'
    const cd = res.headers.get('content-disposition') || 'attachment; filename="consultation.pdf"'
    setHeader(event, 'Content-Type', ct)
    setHeader(event, 'Content-Disposition', cd)
    setHeader(event, 'Cache-Control', 'private, no-store')
    const len = res.headers.get('content-length')
    if (len) setHeader(event, 'Content-Length', len)
    return res._data as ReadableStream
  } catch (e: any) {
    throw createError({
      statusCode: e?.statusCode || e?.response?.status || 502,
      statusMessage: 'download_failed',
    })
  }
})
