import { createError } from 'h3'
import { proxyUpload } from '~/server/utils/api'

/** Aligné sur go/internal/eid.MaxViewerBytes (5 MiB). */
const EID_MAX_UPLOAD_BYTES = 5 << 20

export default defineEventHandler(async (event) => {
  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event, false)
  if (body && body.length > EID_MAX_UPLOAD_BYTES) {
    throw createError({ statusCode: 413, statusMessage: 'eid_file_too_large' })
  }
  return proxyUpload(event, '/api/v1/vet/eid/import', body, contentType)
})
