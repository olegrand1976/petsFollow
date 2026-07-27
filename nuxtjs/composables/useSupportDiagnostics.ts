const WINDOW_MS = 15 * 60 * 1000
const MAX_CONSOLE = 100
const MAX_NETWORK = 50
const ORIGIN_STORAGE_KEY = 'pf_support_origin'

export type SupportConsoleEntry = {
  ts: string
  level: 'error' | 'unhandled' | 'rejection'
  message: string
  stack?: string
  source?: string
}

export type SupportNetworkEntry = {
  ts: string
  method: string
  url: string
  status?: number
  durationMs?: number
  ok?: boolean
}

export type SupportOriginPage = {
  fullPath: string
  path: string
  name: string | null
  params: Record<string, unknown>
  query: Record<string, unknown>
  hash: string
  title: string
  href: string
  capturedAt: string
}

type BufferState = {
  consoleErrors: SupportConsoleEntry[]
  networkEntries: SupportNetworkEntry[]
  installed: boolean
}

const state: BufferState = {
  consoleErrors: [],
  networkEntries: [],
  installed: false,
}

const SENSITIVE_QUERY = /^(token|code|password|access_token|refresh_token|authorization)$/i

function prune<T extends { ts: string }>(items: T[], max: number): T[] {
  const cutoff = Date.now() - WINDOW_MS
  const kept = items.filter((e) => {
    const t = Date.parse(e.ts)
    return Number.isFinite(t) ? t >= cutoff : true
  })
  if (kept.length > max) return kept.slice(kept.length - max)
  return kept
}

function redactUrl(raw: string): string {
  try {
    const u = new URL(raw, typeof window !== 'undefined' ? window.location.origin : 'http://localhost')
    u.searchParams.forEach((_, key) => {
      if (SENSITIVE_QUERY.test(key)) u.searchParams.set(key, '[redacted]')
    })
    if (typeof window !== 'undefined' && u.origin === window.location.origin) {
      return `${u.pathname}${u.search}${u.hash}`
    }
    return u.toString()
  } catch {
    return raw.slice(0, 500)
  }
}

function isSupportPath(path: string): boolean {
  // Exact form page only — do NOT match /admin/support via endsWith.
  const p = (path.split('?')[0] || path).replace(/\/+$/, '') || '/'
  return p === '/support'
}

function redactRecord(input: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {}
  for (const [key, value] of Object.entries(input)) {
    if (SENSITIVE_QUERY.test(key)) {
      out[key] = '[redacted]'
      continue
    }
    out[key] = value
  }
  return out
}

function readOriginFromStorage(): SupportOriginPage | null {
  if (!import.meta.client) return null
  try {
    const raw = sessionStorage.getItem(ORIGIN_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as SupportOriginPage
    if (!parsed?.fullPath || typeof parsed.fullPath !== 'string') return null
    return parsed
  } catch {
    return null
  }
}

function writeOriginToStorage(origin: SupportOriginPage) {
  if (!import.meta.client) return
  try {
    sessionStorage.setItem(ORIGIN_STORAGE_KEY, JSON.stringify(origin))
  } catch {
    // quota / private mode — ignore
  }
}

function clearOriginStorage() {
  if (!import.meta.client) return
  try {
    sessionStorage.removeItem(ORIGIN_STORAGE_KEY)
  } catch {
    // ignore
  }
}

export function peekOriginPage(): SupportOriginPage | null {
  return readOriginFromStorage()
}

export function consumeOriginPage(): SupportOriginPage | null {
  const origin = readOriginFromStorage()
  clearOriginStorage()
  return origin
}

export function pushConsoleError(entry: Omit<SupportConsoleEntry, 'ts'> & { ts?: string }) {
  state.consoleErrors.push({ ts: entry.ts || new Date().toISOString(), ...entry })
  state.consoleErrors = prune(state.consoleErrors, MAX_CONSOLE)
}

export function pushNetworkEntry(entry: Omit<SupportNetworkEntry, 'ts'> & { ts?: string }) {
  state.networkEntries.push({ ts: entry.ts || new Date().toISOString(), ...entry })
  state.networkEntries = prune(state.networkEntries, MAX_NETWORK)
}

function serializeArgs(args: unknown[]): string {
  return args
    .map((a) => {
      if (a instanceof Error) return a.message
      if (typeof a === 'string') return a
      try {
        return JSON.stringify(a)
      } catch {
        return String(a)
      }
    })
    .join(' ')
    .slice(0, 2000)
}

export function installSupportDiagnostics() {
  if (!import.meta.client || state.installed) return
  state.installed = true

  const origError = console.error.bind(console)
  console.error = (...args: unknown[]) => {
    pushConsoleError({
      level: 'error',
      message: serializeArgs(args),
      stack: args.find((a): a is Error => a instanceof Error)?.stack,
    })
    origError(...args)
  }

  window.addEventListener('error', (ev) => {
    pushConsoleError({
      level: 'unhandled',
      message: ev.message || String(ev.error || 'error'),
      stack: ev.error instanceof Error ? ev.error.stack : undefined,
      source: ev.filename ? `${ev.filename}:${ev.lineno}:${ev.colno}` : undefined,
    })
  })

  window.addEventListener('unhandledrejection', (ev) => {
    const reason = ev.reason
    pushConsoleError({
      level: 'rejection',
      message: reason instanceof Error ? reason.message : String(reason),
      stack: reason instanceof Error ? reason.stack : undefined,
    })
  })

  const origFetch = window.fetch.bind(window)
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const method = (init?.method || (input instanceof Request ? input.method : 'GET') || 'GET').toUpperCase()
    const urlRaw = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url
    const started = performance.now()
    const ts = new Date().toISOString()
    try {
      const res = await origFetch(input, init)
      pushNetworkEntry({
        ts,
        method,
        url: redactUrl(urlRaw),
        status: res.status,
        durationMs: Math.round(performance.now() - started),
        ok: res.ok,
      })
      return res
    } catch (err) {
      pushNetworkEntry({
        ts,
        method,
        url: redactUrl(urlRaw),
        durationMs: Math.round(performance.now() - started),
        ok: false,
      })
      throw err
    }
  }
}

export function useSupportDiagnostics() {
  const route = useRoute()
  const { locale } = useI18n()
  const { user } = useProUser()
  const { isDark } = useColorTheme()

  function captureOriginPage() {
    if (!import.meta.client) return
    if (isSupportPath(route.path) || isSupportPath(route.fullPath)) return

    const hrefRaw =
      typeof window !== 'undefined'
        ? `${window.location.pathname}${window.location.search}${window.location.hash}`
        : route.fullPath

    const origin: SupportOriginPage = {
      fullPath: redactUrl(route.fullPath),
      path: route.path,
      name: route.name != null ? String(route.name) : null,
      params: redactRecord({ ...(route.params as Record<string, unknown>) }),
      query: redactRecord({ ...(route.query as Record<string, unknown>) }),
      hash: route.hash || '',
      title: typeof document !== 'undefined' ? document.title : '',
      href: redactUrl(hrefRaw),
      capturedAt: new Date().toISOString(),
    }
    writeOriginToStorage(origin)
  }

  function snapshot() {
    state.consoleErrors = prune(state.consoleErrors, MAX_CONSOLE)
    state.networkEntries = prune(state.networkEntries, MAX_NETWORK)
    const originPage = peekOriginPage()
    const originRoute = originPage?.fullPath || route.fullPath
    return {
      capturedAt: new Date().toISOString(),
      windowMinutes: 15,
      originPage,
      session: {
        userId: user.value?.id || user.value?.userId || null,
        role: user.value?.role || null,
        practiceId: user.value?.practiceId || null,
        practiceName: user.value?.practiceName || null,
      },
      config: {
        route: originRoute,
        locale: locale.value,
        theme: isDark.value ? 'dark' : 'light',
        userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : '',
        appEnv: 'web',
        viewport: typeof window !== 'undefined'
          ? { w: window.innerWidth, h: window.innerHeight }
          : null,
      },
      consoleErrors: [...state.consoleErrors],
      networkEntries: [...state.networkEntries],
    }
  }

  return {
    snapshot,
    captureOriginPage,
    peekOriginPage,
    consumeOriginPage,
    installSupportDiagnostics,
  }
}
