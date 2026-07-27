import { apiBase, authHeaders, localeHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const url = `${apiBase()}/api/v1/vet/pharmacy/batches/export.csv`
  try {
    const res = await $fetch.raw(url, {
      method: 'GET',
      responseType: 'text',
      headers: {
        ...localeHeaders(event),
        ...authHeaders(event),
      },
    })
    setHeader(event, 'Content-Type', res.headers.get('content-type') || 'text/csv; charset=utf-8')
    setHeader(
      event,
      'Content-Disposition',
      res.headers.get('content-disposition') || 'attachment; filename="pharmacy-stock.csv"',
    )
    setHeader(event, 'Cache-Control', 'private, no-store')
    return res._data
  } catch (e: any) {
    throw createError({
      statusCode: e?.statusCode || e?.response?.status || 502,
      statusMessage: 'export_failed',
    })
  }
})
