import type { PacsState } from '~/utils/pacs-poll'
import { pacsPollIntervalMs } from '~/utils/pacs-poll'

export type PacsStatus = {
  state: PacsState
  latencyMs?: number
  checkedAt?: string
  error?: string
}

function unwrap<T>(res: any): T {
  return (res?.data ?? res) as T
}

/**
 * Adaptive PACS status polling: fast while starting, slow when ready/offline.
 */
export function usePacsStatus(opts: { enabled?: Ref<boolean> | boolean } = {}) {
  const status = ref<PacsStatus>({ state: 'offline' })
  const loading = ref(false)
  const waking = ref(false)
  const hasPolled = ref(false)
  const error = ref('')
  let timer: ReturnType<typeof setTimeout> | null = null
  let stopped = false

  const enabled = computed(() => {
    const e = opts.enabled
    if (e == null) return true
    return unref(e)
  })

  function intervalMs(state: PacsState): number {
    return pacsPollIntervalMs(state)
  }

  function clearWakingIfSettled(state: PacsState) {
    // ready = success ; offline = cold start failed / Orthanc down — allow re-wake
    if (state === 'ready' || state === 'offline') waking.value = false
  }

  async function refresh() {
    if (!enabled.value || stopped) return
    loading.value = true
    error.value = ''
    try {
      const res = await $fetch('/api/pacs/status')
      status.value = unwrap<PacsStatus>(res)
      clearWakingIfSettled(status.value.state)
    } catch (e: any) {
      error.value = e?.data?.message || e?.message || 'status_failed'
      status.value = { state: 'offline', error: error.value }
      waking.value = false
    } finally {
      loading.value = false
      hasPolled.value = true
      schedule()
    }
  }

  function schedule() {
    if (timer) clearTimeout(timer)
    if (stopped || !enabled.value) return
    timer = setTimeout(() => { void refresh() }, intervalMs(status.value.state))
  }

  async function wake() {
    if (!enabled.value) return
    waking.value = true
    error.value = ''
    try {
      const res = await $fetch('/api/pacs/wake', { method: 'POST', body: {} })
      status.value = unwrap<PacsStatus>(res)
      schedule()
    } catch (e: any) {
      waking.value = false
      error.value = e?.data?.message || e?.message || 'wake_failed'
    }
  }

  /** Wake then poll until ready or timeout (cold start). */
  async function wakeUntilReady(timeoutMs = 90000): Promise<boolean> {
    await wake()
    const deadline = Date.now() + timeoutMs
    while (Date.now() < deadline) {
      await refresh()
      if (status.value.state === 'ready') return true
      await new Promise((r) => setTimeout(r, 2000))
    }
    const ok = status.value.state === 'ready'
    if (!ok) waking.value = false
    return ok
  }

  function start() {
    stopped = false
    hasPolled.value = false
    void refresh()
  }

  function stop() {
    stopped = true
    if (timer) clearTimeout(timer)
    timer = null
  }

  onMounted(() => {
    if (enabled.value) start()
  })
  onBeforeUnmount(stop)
  watch(enabled, (on) => {
    if (on) start()
    else stop()
  })

  return { status, loading, waking, hasPolled, error, refresh, wake, wakeUntilReady, start, stop }
}
