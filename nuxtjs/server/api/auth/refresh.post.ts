import { apiBase, localeHeaders, absorbAuthTokens, clearAuthCookies } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const refreshToken = getCookie(event, 'pf_refresh')
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
    // Pose les cookies httpOnly et retire les JWT du body renvoyé au navigateur.
    return absorbAuthTokens(event, res)
  } catch (e: any) {
    clearAuthCookies(event)
    throw createError({
      statusCode: e?.statusCode ?? e?.status ?? 401,
      statusMessage: e?.data?.error?.message ?? e?.statusMessage ?? 'Unauthorized',
      data: e?.data,
    })
  }
})
