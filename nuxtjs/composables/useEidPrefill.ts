import { applyEidIdentityToForm, type EidIdentity } from '~/utils/eid-prefill'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

function unwrapData<T>(res: any): T {
  return (res?.data ?? res) as T
}

/** E2E / Vitest : stub sans extension native (cf. `13b-desk-switch` pattern). */
export type PfWebEidMock = {
  status?: () => Promise<{ extension: boolean; nativeApp: boolean }>
  authenticate?: (nonce: string) => Promise<unknown>
}

declare global {
  interface Window {
    __PF_WEB_EID_MOCK__?: PfWebEidMock
  }
}

/** Browser / Vitest with window stub (import.meta.client is false under Node Vitest). */
function isBrowserLike(): boolean {
  return Boolean(import.meta.client) || typeof window !== 'undefined'
}

function getWebEidMock(): PfWebEidMock | null {
  if (!isBrowserLike() || typeof window === 'undefined') return null
  return window.__PF_WEB_EID_MOCK__ ?? null
}

/** Lib Web eID touches `window` at module top-level — never import on SSR. */
async function loadWebEidLibrary() {
  if (!isBrowserLike()) {
    throw new Error('web_eid_client_only')
  }
  const mock = getWebEidMock()
  if (mock) {
    return {
      status: mock.status ?? (async () => ({ extension: true, nativeApp: true })),
      authenticate: mock.authenticate ?? (async () => ({ mocked: true })),
    }
  }
  return import('@web-eid/web-eid-library')
}

function isWebEidUnavailableError(e: unknown): boolean {
  const msg = String((e as any)?.message || e || '').toLowerCase()
  const code = String((e as any)?.code || (e as any)?.name || '').toLowerCase()
  return (
    code.includes('extensionunavailable')
    || code.includes('nativeappunavailable')
    || msg.includes('extension')
    || msg.includes('native app')
    || msg.includes('web-eid')
    || msg.includes('webeid')
  )
}

export function useEidPrefill() {
  const config = useRuntimeConfig()
  const enabled = computed(() => isPublicFlagOn(config.public.eidEnabled))

  async function checkWebEidStatus(): Promise<
    | { ok: true; hasExtension: boolean; hasNativeApp: boolean }
    | { ok: false; message: string }
  > {
    if (!isBrowserLike()) {
      return { ok: false, message: 'web_eid_client_only' }
    }
    try {
      const { status } = await loadWebEidLibrary()
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
    if (!isBrowserLike()) {
      throw new Error('web_eid_client_only')
    }
    const challengeRes: any = await $fetch('/api/vet/eid/web-eid/challenge')
    const challenge = unwrapData<{ nonce: string; origin?: string }>(challengeRes)
    if (!challenge?.nonce) {
      throw new Error('eid_nonce_missing')
    }
    // Mismatch site origin ↔ tab origin fails late in the extension — fail fast with a clear code.
    if (challenge.origin && typeof window !== 'undefined') {
      const pageOrigin = window.location.origin
      if (challenge.origin !== pageOrigin) {
        throw new Error(`eid_origin_mismatch:${challenge.origin}!=${pageOrigin}`)
      }
    }
    let token: unknown
    try {
      const { authenticate } = await loadWebEidLibrary()
      token = await authenticate(challenge.nonce)
    } catch (e) {
      if (isWebEidUnavailableError(e)) {
        throw new Error('web_eid_unavailable')
      }
      throw e
    }
    const verifyRes: any = await $fetch('/api/vet/eid/web-eid/verify', {
      method: 'POST',
      body: { token },
    })
    return unwrapData<EidIdentity>(verifyRes)
  }

  return {
    enabled,
    checkWebEidStatus,
    uploadViewerFile,
    readCard,
    applyToForm: applyEidIdentityToForm,
  }
}
