import { apiBase, localeHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getRouterParam(event, 'token')
  const url = `${apiBase()}/api/v1/public/pet-dossier/${encodeURIComponent(token || '')}/download`
  try {
    const res = await $fetch.raw(url, {
      method: 'GET',
      responseType: 'arrayBuffer',
      headers: localeHeaders(event),
    })
    const ct = res.headers.get('content-type') || 'application/zip'
    const cd = res.headers.get('content-disposition') || 'attachment; filename="dossier.zip"'
    setHeader(event, 'Content-Type', ct)
    setHeader(event, 'Content-Disposition', cd)
    setHeader(event, 'Cache-Control', 'private, no-store')
    return Buffer.from(res._data as ArrayBuffer)
  } catch (e: any) {
    const status = e?.statusCode || e?.response?.status || 502
    throw createError({
      statusCode: status,
      statusMessage: e?.statusMessage || 'download_failed',
      data: e?.data,
    })
  }
})
