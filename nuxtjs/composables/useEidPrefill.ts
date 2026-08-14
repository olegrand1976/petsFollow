import { applyEidIdentityToForm, type EidIdentity } from '~/utils/eid-prefill'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

/** UI hard-stop if native messaging never ACKs / hangs beyond library timeouts. */
const WEB_EID_AUTH_TIMEOUT_MS = 90_000

function unwrapData<T>(res: any): T {
  return (res?.data ?? res) as T
}

/** E2E / Vitest : stub sans extension native (cf. `13b-desk-switch` pattern). */
export type PfWebEidMock = {
  status?: () => Promise<{ extension: boolean | string; nativeApp: boolean | string }>
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
      status: mock.status ?? (async () => ({ extension: '2.0.0', nativeApp: '2.0.0' })),
      authenticate: mock.authenticate ?? (async () => ({ mocked: true })),
    }
  }
  return import('@web-eid/web-eid-library')
}

function errorBag(e: unknown): string {
  const any = e as any
  return [
    any?.code,
    any?.name,
    any?.message,
    any?.cause?.code,
    any?.cause?.name,
    any?.cause?.message,
  ]
    .filter(Boolean)
    .join(' ')
    .toLowerCase()
}

function isLoopbackHttpOrigin(origin: string): boolean {
  try {
    const u = new URL(origin)
    if (u.protocol !== 'http:') return false
    const h = u.hostname.toLowerCase()
    return h === 'localhost' || h === '127.0.0.1' || h === '[::1]' || h === '::1'
  } catch {
    return false
  }
}

/**
 * Map @web-eid/web-eid-library errors to stable UI codes.
 * ExtensionUnavailable = no postMessage ACK in 1s (extension missing/disabled/wrong browser).
 */
export function mapWebEidThrownError(e: unknown): string {
  const blob = errorBag(e)
  if (blob.includes('web_eid_timeout') || blob.includes('err_webeid_action_timeout')) {
    return 'web_eid_timeout'
  }
  if (blob.includes('err_webeid_user_timeout') || blob.includes('usertimeouterror')) {
    return 'web_eid_timeout'
  }
  if (blob.includes('err_webeid_user_cancelled') || blob.includes('usercancelled')) {
    return 'web_eid_cancelled'
  }
  if (blob.includes('err_webeid_context_insecure') || blob.includes('contextinsecure')) {
    return 'web_eid_context_insecure'
  }
  if (blob.includes('err_webeid_version_mismatch') || blob.includes('versionmismatch')) {
    return 'web_eid_version_mismatch'
  }
  if (blob.includes('err_webeid_native_unavailable') || blob.includes('nativeunavailable')) {
    return 'web_eid_native_unavailable'
  }
  if (
    blob.includes('err_webeid_extension_unavailable')
    || blob.includes('extensionunavailable')
    || blob.includes('extension is not available')
  ) {
    return 'web_eid_extension_unavailable'
  }
  if (blob.includes('err_webeid_native_fatal') || blob.includes('nativefatal')) {
    return 'web_eid_unavailable'
  }
  return ''
}

async function withTimeout<T>(promise: Promise<T>, ms: number, code: string): Promise<T> {
  let timer: ReturnType<typeof setTimeout> | undefined
  try {
    return await Promise.race([
      promise,
      new Promise<never>((_, reject) => {
        timer = setTimeout(() => reject(new Error(code)), ms)
      }),
    ])
  } finally {
    if (timer) clearTimeout(timer)
  }
}

export function useEidPrefill() {
  const config = useRuntimeConfig()
  const enabled = computed(() => isPublicFlagOn(config.public.eidEnabled))

  async function checkWebEidStatus(): Promise<
    | { ok: true; hasExtension: boolean; hasNativeApp: boolean; detail?: string }
    | { ok: false; message: string; code?: string }
  > {
    if (!isBrowserLike()) {
      return { ok: false, message: 'web_eid_client_only' }
    }
    try {
      const { status } = await loadWebEidLibrary()
      const st = await status()
      // Library returns semver strings when present.
      const hasExtension = Boolean(st?.extension)
      const hasNativeApp = Boolean(st?.nativeApp)
      return {
        ok: true,
        hasExtension,
        hasNativeApp,
        detail: `ext=${st?.extension || '-'} app=${st?.nativeApp || '-'}`,
      }
    } catch (e: any) {
      const code = mapWebEidThrownError(e) || 'web_eid_unavailable'
      return { ok: false, message: e?.message || String(e), code }
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

    const pageOrigin = typeof window !== 'undefined' ? window.location.origin : ''
    const onLoopback = pageOrigin ? isLoopbackHttpOrigin(pageOrigin) : false

    // On HTTPS (staging/prod): fail fast if extension/app missing.
    // On localhost/127.0.0.1: many browsers do not inject Web eID → status() false-negatives
    // even when the same install works on staging — skip hard gate and try authenticate.
    const st = await checkWebEidStatus()
    if (!onLoopback) {
      if (st.ok === false) {
        throw new Error(st.code || 'web_eid_unavailable', { cause: new Error(st.message) })
      }
      if (!st.hasExtension) {
        throw new Error('web_eid_extension_unavailable')
      }
      if (!st.hasNativeApp) {
        throw new Error('web_eid_native_unavailable')
      }
    }

    // Fresh challenge immediately before authenticate (PIN dialog can take a while).
    const challengeRes: any = await $fetch('/api/vet/eid/web-eid/challenge')
    const challenge = unwrapData<{ nonce: string; origin?: string }>(challengeRes)
    if (!challenge?.nonce) {
      throw new Error('eid_nonce_missing')
    }
    if (challenge.origin && pageOrigin && challenge.origin !== pageOrigin) {
      throw new Error(`eid_origin_mismatch:${challenge.origin}!=${pageOrigin}`)
    }

    let token: unknown
    try {
      const { authenticate } = await loadWebEidLibrary()
      const authPromise = Promise.resolve(
        authenticate(challenge.nonce, {
          userInteractionTimeout: WEB_EID_AUTH_TIMEOUT_MS,
          lang: typeof navigator !== 'undefined' ? navigator.language?.slice(0, 2) : 'fr',
        } as { userInteractionTimeout?: number; lang?: string }),
      )
      token = await withTimeout(authPromise, WEB_EID_AUTH_TIMEOUT_MS + 10_000, 'web_eid_timeout')
    } catch (e) {
      const mapped = mapWebEidThrownError(e)
      if (onLoopback && (
        mapped === 'web_eid_extension_unavailable'
        || mapped === 'web_eid_unavailable'
        || mapped === 'web_eid_native_unavailable'
      )) {
        throw new Error('web_eid_loopback_unavailable', { cause: e })
      }
      if (mapped) {
        throw new Error(mapped, { cause: e })
      }
      if ((e as any)?.message === 'web_eid_timeout') {
        throw new Error('web_eid_timeout', { cause: e })
      }
      throw e
    }
    try {
      const verifyRes: any = await $fetch('/api/vet/eid/web-eid/verify', {
        method: 'POST',
        body: { token },
      })
      return unwrapData<EidIdentity>(verifyRes)
    } catch (e: any) {
      const msgKey = e?.data?.data?.error?.msgKey || e?.data?.error?.msgKey || e?.data?.error?.messageKey
      if (typeof msgKey === 'string' && msgKey.startsWith('eid_')) {
        throw new Error(msgKey, { cause: e })
      }
      throw e
    }
  }

  return {
    enabled,
    checkWebEidStatus,
    uploadViewerFile,
    readCard,
    applyToForm: applyEidIdentityToForm,
  }
}
