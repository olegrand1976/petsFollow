import { buildCsp } from '~/utils/buildCsp'
import { proxyBinary } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'id required' })
  }
  // Filet si routeRules merge mal : Firefox refuse l'iframe avec DENY / frame-ancestors none.
  setHeader(event, 'X-Frame-Options', 'SAMEORIGIN')
  setHeader(event, 'Content-Security-Policy', buildCsp(undefined, { frameAncestors: "'self'" }))
  return proxyBinary(event, `/api/v1/admin/compendium-imports/${id}/pdf`, {
    contentDispositionFallback: 'inline; filename="compendium.pdf"',
  })
})
