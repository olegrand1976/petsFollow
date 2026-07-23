export type ProUser = {
  id?: string
  userId?: string
  email?: string
  fullName?: string
  avatarUrl?: string
  role?: string
  practiceId?: string
  practiceName?: string
  emailVerified?: boolean
  profileComplete?: boolean
  preferredLocale?: string
  mustChangePassword?: boolean
}

export function useProUser() {
  const userState = useState<ProUser | null>('pro-user', () => null)
  const loadingState = useState<boolean>('pro-user-loading', () => false)

  const user = computed(() => userState.value)
  const loading = computed(() => loadingState.value)

  async function fetchUser(force = false) {
    if (userState.value && !force) return userState.value
    loadingState.value = true
    try {
      const res: any = await $fetch('/api/me')
      const data = res.data ?? res
      userState.value = data
      return data as ProUser
    } catch (e: any) {
      const status = e?.statusCode ?? e?.status ?? e?.response?.status
      if (status === 401 || status === 403) {
        userState.value = null
        throw e
      }
      // Transient (5xx/network): keep cached profile when available.
      return userState.value
    } finally {
      loadingState.value = false
    }
  }

  function initials() {
    const name = userState.value?.fullName || userState.value?.email || '?'
    return name
      .split(' ')
      .map((p) => p[0])
      .join('')
      .slice(0, 2)
      .toUpperCase()
  }

  function logout() {
    clearAuthTokens()
    userState.value = null
    navigateTo('/login')
  }

  return { user, loading, fetchUser, initials, logout }
}
