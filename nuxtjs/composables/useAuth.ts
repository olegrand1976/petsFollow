export type AuthTokens = {
  accessToken: string
  refreshToken?: string
  expiresIn?: number
}

export type AuthMFAChallenge = {
  requires2FA: true
  mfaToken: string
  expiresIn?: number
}

export type AuthResponse = AuthTokens | AuthMFAChallenge

export function isMFAChallenge(res: AuthResponse): res is AuthMFAChallenge {
  return 'requires2FA' in res && res.requires2FA === true
}

export function parseJwtRole(token: string | null | undefined): string | null {
  if (!token) return null
  const parts = token.split('.')
  if (parts.length < 2) return null
  try {
    const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')))
    return payload.role as string
  } catch {
    return null
  }
}

export function unwrapAuthData(res: unknown): AuthResponse {
  const data = (res as { data?: AuthResponse })?.data ?? res
  return data as AuthResponse
}

export function extractAccessToken(res: AuthResponse): string | null {
  if (isMFAChallenge(res)) return null
  return res.accessToken ?? null
}
