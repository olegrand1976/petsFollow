import { apiBase, authHeaders, localeHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'missing_sim_id' })
  }
  const url = `${apiBase()}/api/v1/commercial/pitch-sims/${encodeURIComponent(id)}/audio`
  try {
    const res = await $fetch.raw(url, {
      method: 'GET',
      responseType: 'arrayBuffer',
      headers: { ...localeHeaders(event), ...authHeaders(event) },
    })
    setHeader(event, 'Content-Type', res.headers.get('content-type') || 'audio/webm')
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
