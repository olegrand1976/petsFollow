import { proxyApi } from '~/server/utils/api'

export default defineEventHandler(async (event) => {
  const hasToken = !!getCookie(event, 'pf_token')
  const hasRefresh = !!getCookie(event, 'pf_refresh')
  const hasSession = !!getCookie(event, 'pf_session')

  if (!hasToken && !hasRefresh) {
    // Ne jamais logger les valeurs des cookies (JWT).
    console.warn('[auth/me] 401 missing_cookies', {
      reason: 'missing_cookies',
      hasToken,
      hasRefresh,
      hasSession,
    })
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized',
      data: { reason: 'missing_cookies' },
    })
  }

  try {
    return await proxyApi(event, '/api/v1/me')
  } catch (e: unknown) {
    const err = e as {
      statusCode?: number
      status?: number
      data?: { error?: { code?: string }, data?: { error?: { code?: string } }, reason?: string }
    }
    const status = err?.statusCode ?? err?.status ?? 500
    if (status === 401 || status === 403) {
      console.warn('[auth/me] 401 upstream', {
        reason: 'upstream_unauthorized',
        hasToken,
        hasRefresh,
        hasSession,
        code: err?.data?.error?.code ?? err?.data?.data?.error?.code ?? null,
      })
    }
    throw e
  }
})
