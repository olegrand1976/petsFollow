import { proxyApi } from '~/server/utils/api'

function upstreamErrorCode(e: unknown): string | null {
  const err = e as {
    data?: { error?: { code?: string }, data?: { error?: { code?: string } } }
  }
  return err?.data?.error?.code ?? err?.data?.data?.error?.code ?? null
}

export default defineEventHandler(async (event) => {
  const hasToken = !!getCookie(event, 'pf_token')
  const hasRefresh = !!getCookie(event, 'pf_refresh')
  const hasSession = !!getCookie(event, 'pf_session')
  const cookieHeaderPresent = !!getHeader(event, 'cookie')

  if (!hasToken && !hasRefresh) {
    // Ne jamais logger les valeurs des cookies (JWT).
    console.warn('[auth/me] 401 missing_cookies', {
      reason: 'missing_cookies',
      hasToken,
      hasRefresh,
      hasSession,
      cookieHeaderPresent,
    })
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized',
      data: {
        reason: 'missing_cookies',
        hasSession,
        cookieHeaderPresent,
      },
    })
  }

  try {
    return await proxyApi(event, '/api/v1/me')
  } catch (e: unknown) {
    const err = e as { statusCode?: number, status?: number }
    const status = err?.statusCode ?? err?.status ?? 500
    if (status === 401 || status === 403) {
      const code = upstreamErrorCode(e)
      // invalid_token = signature/typ/exp côté Go ; missing_bearer = BFF n'a pas forwardé.
      console.warn('[auth/me] 401 upstream', {
        reason: code === 'invalid_token' ? 'invalid_token' : 'upstream_unauthorized',
        hasToken,
        hasRefresh,
        hasSession,
        cookieHeaderPresent,
        code,
      })
    }
    throw e
  }
})
