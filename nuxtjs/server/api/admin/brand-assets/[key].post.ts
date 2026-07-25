import { proxyUpload } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const key = getRouterParam(event, 'key')
  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event, false)
  return proxyUpload(event, `/api/v1/admin/brand-assets/${encodeURIComponent(key || '')}`, body, contentType)
})
