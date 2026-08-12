import { apiBase, localeHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getRouterParam(event, 'token')
  const url = `${apiBase()}/api/v1/public/pet-dossier/${encodeURIComponent(token || '')}/download`
  try {
    // Streaming plutôt qu'arrayBuffer : le ZIP contient tout le dossier médical et
    // n'a pas à être bufferisé une seconde fois en RAM par la BFF.
    const res = await $fetch.raw(url, {
      method: 'GET',
      responseType: 'stream',
      headers: localeHeaders(event),
    })
    const ct = res.headers.get('content-type') || 'application/zip'
    const cd = res.headers.get('content-disposition') || 'attachment; filename="dossier.zip"'
    setHeader(event, 'Content-Type', ct)
    setHeader(event, 'Content-Disposition', cd)
    setHeader(event, 'Cache-Control', 'private, no-store')
    // Sans ce relais, le passage en flux prive le navigateur de la progression
    // du téléchargement, que la version bufferisée fournissait implicitement.
    const len = res.headers.get('content-length')
    const n = len ? Number.parseInt(len, 10) : NaN
    if (Number.isFinite(n) && n >= 0) setHeader(event, 'Content-Length', n)
    return res._data as ReadableStream
  } catch (e: any) {
    // En mode stream, le corps d'erreur amont est lui aussi un flux : on ne
    // renvoie que le statut, sinon createError échoue à le sérialiser.
    throw createError({
      statusCode: e?.statusCode || e?.response?.status || 502,
      statusMessage: 'download_failed',
    })
  }
})
