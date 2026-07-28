import { apiBase, authHeaders, localeHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const visitId = getRouterParam(event, 'visitId')
  if (!visitId) {
    throw createError({ statusCode: 400, statusMessage: 'missing_visit_id' })
  }
  const url = `${apiBase()}/api/v1/visits/${encodeURIComponent(visitId)}/report/audio`
  try {
    const res = await $fetch.raw(url, {
      method: 'GET',
      responseType: 'arrayBuffer',
      headers: { ...localeHeaders(event), ...authHeaders(event) },
    })
    setHeader(event, 'Content-Type', res.headers.get('content-type') || 'audio/mpeg')
    setHeader(event, 'Cache-Control', 'private, no-store')
    return res._data
  }
  catch (e: any) {
    throw createError({
      statusCode: e?.statusCode || e?.response?.status || 502,
      statusMessage: e?.statusMessage || e?.data?.error?.msgKey || 'audio_failed',
    })
  }
})
