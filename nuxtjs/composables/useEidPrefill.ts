import { applyEidIdentityToForm, type EidFormTarget, type EidIdentity } from '~/utils/eid-prefill'

export type { EidFormTarget, EidIdentity }

const KNOWN_EID_EXTENSIONS = Object.freeze({
  beidconnect: 'pencgnkbgaekikmiahiaakjdgaibiipp',
  eid_chrome: 'bkbdaodnaecdijpajecpncpdomgcoakc',
})

const PASSIVE_DETECT_TIMEOUT_MS = 800

function unwrapData<T>(res: any): T {
  return (res?.data ?? res) as T
}

function detectChromeExtension(extensionId: string): Promise<boolean> {
  const chromeApi = (globalThis as any).chrome
  if (!chromeApi?.runtime?.sendMessage) {
    return Promise.resolve(false)
  }
  return new Promise((resolve) => {
    let settled = false
    const finish = (value: boolean) => {
      if (settled) return
      settled = true
      resolve(value)
    }
    const timer = setTimeout(() => finish(false), PASSIVE_DETECT_TIMEOUT_MS)
    try {
      chromeApi.runtime.sendMessage(extensionId, { ping: true }, () => {
        clearTimeout(timer)
        const lastError = chromeApi.runtime.lastError
        if (!lastError) {
          finish(true)
          return
        }
        const msg = String(lastError.message || '')
        if (/Could not establish connection|message port closed|did not respond|disconnected/i.test(msg)) {
          finish(true)
          return
        }
        finish(false)
      })
    } catch {
      clearTimeout(timer)
      finish(false)
    }
  })
}

export function useEidPrefill() {
  const config = useRuntimeConfig()
  const enabled = computed(() => Boolean(config.public.eidEnabled))

  async function detectInstalledEidExtensions() {
    const [beidconnect, eidChrome] = await Promise.all([
      detectChromeExtension(KNOWN_EID_EXTENSIONS.beidconnect),
      detectChromeExtension(KNOWN_EID_EXTENSIONS.eid_chrome),
    ])
    return { beidconnect, eid_chrome: eidChrome }
  }

  async function checkWebEidStatus(): Promise<
    | { ok: true; hasExtension: boolean; hasNativeApp: boolean }
    | { ok: false; message: string }
  > {
    try {
      const { status } = await import('@web-eid/web-eid-library')
      const { extension, nativeApp } = await status()
      const hasExtension = Boolean(extension)
      const hasNativeApp = Boolean(nativeApp)
      return { ok: true, hasExtension, hasNativeApp }
    } catch (e: any) {
      return { ok: false, message: e?.message || String(e) }
    }
  }

  async function uploadViewerFile(file: File): Promise<EidIdentity> {
    const form = new FormData()
    form.append('file', file)
    const res = await $fetch('/api/vet/eid/import', { method: 'POST', body: form })
    return unwrapData<EidIdentity>(res)
  }

  async function readCard(): Promise<EidIdentity> {
    const challengeRes: any = await $fetch('/api/vet/eid/web-eid/challenge')
    const challenge = unwrapData<{ nonce: string; origin?: string }>(challengeRes)
    if (!challenge?.nonce) {
      throw new Error('eid_nonce_missing')
    }
    const { authenticate } = await import('@web-eid/web-eid-library')
    const token = await authenticate(challenge.nonce)
    const verifyRes: any = await $fetch('/api/vet/eid/web-eid/verify', {
      method: 'POST',
      body: { token },
    })
    return unwrapData<EidIdentity>(verifyRes)
  }

  return {
    enabled,
    detectInstalledEidExtensions,
    checkWebEidStatus,
    uploadViewerFile,
    readCard,
    applyToForm: applyEidIdentityToForm,
  }
}
