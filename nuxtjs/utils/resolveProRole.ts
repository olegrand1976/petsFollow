/**
 * Resolve the current Pro role for route guards.
 * Prefers JWT claim; falls back to /api/me when only a refresh cookie is present.
 */
export async function resolveProRole(): Promise<string | null> {
  const token = useCookie('pf_token')
  const fromJwt = parseJwtRole(token.value)
  if (fromJwt) return fromJwt

  const refresh = useCookie('pf_refresh')
  if (!token.value && !refresh.value) return null

  try {
    const { fetchUser } = useProUser()
    const me = await fetchUser(true)
    return me?.role ?? null
  } catch {
    return null
  }
}
