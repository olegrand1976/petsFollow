export type PacsDebugEntry = {
  at: string
  method: string
  url: string
  ok: boolean
  ms: number
  status?: number
  message?: string
}

const MAX = 80

/**
 * Ring buffer for admin PACS playground client-side debug (fetch traces).
 */
export function usePacsDebugLog() {
  const entries = useState<PacsDebugEntry[]>('pacs-debug-log', () => [])

  function push(entry: PacsDebugEntry) {
    entries.value = [entry, ...entries.value].slice(0, MAX)
    if (import.meta.client) {
      console.debug('[pacs]', entry.ok ? 'ok' : 'err', entry.method, entry.url, `${entry.ms}ms`, entry.message || '')
    }
  }

  function clear() {
    entries.value = []
  }

  return { entries, push, clear }
}

/** Instrumented $fetch for PacsViewerContainer debug mode. */
export async function pacsDebugFetch<T = any>(
  url: string,
  opts: Record<string, any> | undefined,
  onDebug?: (e: PacsDebugEntry) => void,
): Promise<T> {
  const method = String(opts?.method || 'GET').toUpperCase()
  const started = Date.now()
  const isStatusPoll = method === 'GET' && url.split('?')[0] === '/api/pacs/status'
  try {
    const res = await ($fetch as any)(url, opts) as T
    // Skip successful status polls — they flood the admin debug ring every few seconds.
    if (!isStatusPoll) {
      onDebug?.({
        at: new Date().toISOString(),
        method,
        url,
        ok: true,
        ms: Date.now() - started,
        status: 200,
      })
    }
    return res
  } catch (e: any) {
    onDebug?.({
      at: new Date().toISOString(),
      method,
      url,
      ok: false,
      ms: Date.now() - started,
      status: e?.statusCode || e?.response?.status || e?.status,
      message: e?.data?.message || e?.data?.error || e?.message,
    })
    throw e
  }
}
