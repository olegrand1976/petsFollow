import { apiBase, localeHeaders, setAuthCookies, clearAuthCookies } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const body = await readBody<{ refreshToken?: string }>(event).catch(() => ({} as { refreshToken?: string }))
  const refreshToken = body?.refreshToken || getCookie(event, 'pf_refresh')
  if (!refreshToken) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  try {
    const res: any = await $fetch(`${apiBase()}/api/v1/auth/refresh`, {
      method: 'POST',
      body: { refreshToken },
      headers: localeHeaders(event),
    })
    const pair = res.data ?? res
    if (!pair?.accessToken) {
      throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
    }
    setAuthCookies(event, pair)
    return res
  } catch (e: any) {
    clearAuthCookies(event)
    throw createError({
      statusCode: e?.statusCode ?? e?.status ?? 401,
      statusMessage: e?.data?.error?.message ?? e?.statusMessage ?? 'Unauthorized',
      data: e?.data,
    })
  }
})
