import { apiBase, authHeaders, localeHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  const url = `${apiBase()}/api/v1/vet/prescriptions/${id}/pdf`
  try {
    const res = await $fetch.raw(url, {
      method: 'GET',
      responseType: 'arrayBuffer',
      headers: { ...localeHeaders(event), ...authHeaders(event) },
    })
    setHeader(event, 'Content-Type', res.headers.get('content-type') || 'application/pdf')
    setHeader(event, 'Content-Disposition', res.headers.get('content-disposition') || 'inline; filename="prescription.pdf"')
    setHeader(event, 'Cache-Control', 'private, no-store')
    return res._data
  } catch (e: any) {
    throw createError({
      statusCode: e?.statusCode || e?.response?.status || 502,
      statusMessage: 'pdf_failed',
    })
  }
})
