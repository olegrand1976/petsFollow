const WINDOW_MS = 15 * 60 * 1000
const MAX_CONSOLE = 100
const MAX_NETWORK = 50

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
      return `${u.pathname}${u.search}`
    }
    return u.toString()
  } catch {
    return raw.slice(0, 500)
  }
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

  function snapshot() {
    state.consoleErrors = prune(state.consoleErrors, MAX_CONSOLE)
    state.networkEntries = prune(state.networkEntries, MAX_NETWORK)
    return {
      capturedAt: new Date().toISOString(),
      windowMinutes: 15,
      session: {
        userId: user.value?.id || user.value?.userId || null,
        role: user.value?.role || null,
        practiceId: user.value?.practiceId || null,
        practiceName: user.value?.practiceName || null,
      },
      config: {
        route: route.fullPath,
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

  return { snapshot, installSupportDiagnostics }
}
