import { authErrorStatus, clearAuthTokens, isWithinPostLoginGrace } from '~/composables/useAuth'

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
  isReferenceVet?: boolean
}

type InFlight = {
  force: boolean
  promise: Promise<ProUser | null>
}

export function useProUser() {
  const userState = useState<ProUser | null>('pro-user', () => null)
  const loadingState = useState<boolean>('pro-user-loading', () => false)
  const inflightState = useState<InFlight | null>('pro-user-inflight', () => null)

  const user = computed(() => userState.value)
  const loading = computed(() => loadingState.value)

  async function fetchUserOnce(force: boolean, allowGraceRetry: boolean): Promise<ProUser | null> {
    loadingState.value = true
    try {
      // SSR: never $fetch our own /api/* via the public URL — on Cloud Run that
      // re-enters the same instance and deadlocks (concurrency slot) then OOMs.
      // useRequestFetch = appel Nitro interne (cookies forwardés, refresh BFF OK).
      const res: any = import.meta.server
        ? await useRequestFetch()('/api/me')
        : await $fetch('/api/me')
      const data = res.data ?? res
      userState.value = data
      return data as ProUser
    } catch (e: unknown) {
      const status = authErrorStatus(e)
      if (status === 401 || status === 403) {
        // Une seule reprise pendant la grâce post-login (cookies WebKit pas encore attachés).
        if (allowGraceRetry && isWithinPostLoginGrace()) {
          await new Promise((r) => setTimeout(r, 200))
          return fetchUserOnce(true, false)
        }
        userState.value = null
        throw e
      }
      // Transient (5xx/network): keep cached profile when available.
      return userState.value
    } finally {
      loadingState.value = false
    }
  }

  async function fetchUser(force = false) {
    if (userState.value && !force) return userState.value

    const inflight = inflightState.value
    // Coalesce concurrent callers (middlewares + layout) into one /api/me.
    if (inflight && (!force || inflight.force)) {
      return inflight.promise
    }

    const promise = fetchUserOnce(force, true)
    inflightState.value = { force, promise }
    try {
      return await promise
    } finally {
      if (inflightState.value?.promise === promise) {
        inflightState.value = null
      }
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

  async function logout() {
    await clearAuthTokens()
    userState.value = null
    await navigateTo('/login')
  }

  return { user, loading, fetchUser, initials, logout }
}
