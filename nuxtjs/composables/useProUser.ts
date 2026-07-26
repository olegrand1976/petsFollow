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
  gen: number
  promise: Promise<ProUser | null>
}

export function useProUser() {
  const userState = useState<ProUser | null>('pro-user', () => null)
  const loadingState = useState<boolean>('pro-user-loading', () => false)
  const inflightState = useState<InFlight | null>('pro-user-inflight', () => null)
  const fetchGen = useState<number>('pro-user-fetch-gen', () => 0)

  const user = computed(() => userState.value)
  const loading = computed(() => loadingState.value)

  async function fetchUser(force = false) {
    if (userState.value && !force) return userState.value

    const inflight = inflightState.value
    // Coalesce concurrent callers (middlewares + layout) into one /api/me.
    if (inflight && (!force || inflight.force)) {
      return inflight.promise
    }

    const gen = fetchGen.value + 1
    fetchGen.value = gen
    loadingState.value = true

    const settled = (async () => {
      const runOnce = async (allowGraceRetry: boolean): Promise<ProUser | null> => {
        try {
          // SSR: never $fetch our own /api/* via the public URL — on Cloud Run that
          // re-enters the same instance and deadlocks (concurrency slot) then OOMs.
          // useRequestFetch = appel Nitro interne (cookies forwardés, refresh BFF OK).
          const res: any = import.meta.server
            ? await useRequestFetch()('/api/me')
            : await $fetch('/api/me')
          if (gen !== fetchGen.value) return userState.value
          const data = res.data ?? res
          userState.value = data
          return data as ProUser
        } catch (e: unknown) {
          const status = authErrorStatus(e)
          if (status === 401 || status === 403) {
            // Une seule reprise pendant la grâce post-login (client only ; SSR n'a pas sessionStorage).
            if (allowGraceRetry && isWithinPostLoginGrace()) {
              await new Promise((r) => setTimeout(r, 200))
              return await runOnce(false)
            }
            if (gen === fetchGen.value) userState.value = null
            throw e
          }
          // Transient (5xx/network): keep cached profile when available.
          return userState.value
        }
      }

      try {
        return await runOnce(true)
      } finally {
        if (gen === fetchGen.value) {
          loadingState.value = false
          if (inflightState.value?.gen === gen) {
            inflightState.value = null
          }
        }
      }
    })()

    inflightState.value = { force, gen, promise: settled }
    return settled
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
