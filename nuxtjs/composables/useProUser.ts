export type ProUser = {
  id?: string
  userId?: string
  email?: string
  fullName?: string
  role?: string
  practiceId?: string
  practiceName?: string
  emailVerified?: boolean
  profileComplete?: boolean
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
    } catch {
      userState.value = null
      return null
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
    const token = useCookie('pf_token')
    token.value = null
    userState.value = null
    navigateTo('/login')
  }

  return { user, loading, fetchUser, initials, logout }
}
