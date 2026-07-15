import { apiBase, apiHeaders } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const token = getCookie(event, 'pf_token')
  if (!token) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  const petId = getRouterParam(event, 'petId')
  if (!petId) {
    throw createError({ statusCode: 400, statusMessage: 'petId required' })
  }

  const contentType = getHeader(event, 'content-type') || ''
  const body = await readRawBody(event)
  try {
    return await $fetch(`${apiBase()}/api/v1/pets/${petId}/photo`, {
      method: 'POST',
      body,
      headers: {
        ...apiHeaders(event),
        'content-type': contentType,
      },
    })
  } catch (e: any) {
    throw createError({
      statusCode: e?.statusCode ?? e?.status ?? 500,
      statusMessage: e?.data?.error?.message ?? e?.statusMessage ?? 'Error',
      data: e?.data,
    })
  }
})
