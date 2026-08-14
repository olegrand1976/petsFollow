import { beforeEach, describe, expect, it, vi } from 'vitest'

const fetchMock = vi.fn()

vi.stubGlobal('$fetch', fetchMock)
vi.stubGlobal('computed', <T>(fn: () => T) => ({
  get value() {
    return fn()
  },
}))
vi.stubGlobal('useRuntimeConfig', () => ({ public: { eidEnabled: true } }))
vi.stubGlobal('import.meta', { client: true })

describe('useEidPrefill readCard', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    vi.resetModules()
    vi.stubGlobal('$fetch', fetchMock)
    vi.stubGlobal('computed', <T>(fn: () => T) => ({
      get value() {
        return fn()
      },
    }))
    vi.stubGlobal('useRuntimeConfig', () => ({ public: { eidEnabled: true } }))
    // jsdom-like window for origin check + mock hook
    const win = {
      location: { origin: 'http://localhost:3002' },
      __PF_WEB_EID_MOCK__: {
        authenticate: async () => ({ unverifiedCertificate: 'unit' }),
        status: async () => ({ extension: true, nativeApp: true }),
      },
    }
    vi.stubGlobal('window', win)
    vi.stubGlobal('import.meta', { client: true })
  })

  it('challenge → authenticate → verify returns identity', async () => {
    fetchMock
      .mockResolvedValueOnce({
        data: { nonce: 'n1', origin: 'http://localhost:3002' },
      })
      .mockResolvedValueOnce({
        data: {
          firstname: 'Camille',
          lastname: 'Testeur',
          niss: '96072399886',
          import_tool: 'web_eid',
        },
      })

    const win = window as Window & { __PF_WEB_EID_MOCK__?: unknown }
    win.__PF_WEB_EID_MOCK__ = {
      status: async () => ({ extension: '2.0.0', nativeApp: '2.0.0' }),
      authenticate: async () => ({ unverifiedCertificate: 'unit' }),
    }

    const { useEidPrefill } = await import('../../composables/useEidPrefill')
    const { readCard } = useEidPrefill()
    const identity = await readCard()
    expect(identity.firstname).toBe('Camille')
    expect(identity.niss).toBe('96072399886')
    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/vet/eid/web-eid/challenge')
    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/vet/eid/web-eid/verify', {
      method: 'POST',
      body: { token: { unverifiedCertificate: 'unit' } },
    })
  })

  it('fails fast on origin mismatch', async () => {
    fetchMock.mockResolvedValueOnce({
      data: { nonce: 'n1', origin: 'https://evil.example' },
    })
    const { useEidPrefill } = await import('../../composables/useEidPrefill')
    const { readCard } = useEidPrefill()
    await expect(readCard()).rejects.toThrow(/eid_origin_mismatch/)
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('maps user_timeout from authenticate', async () => {
    const win = window as Window & { __PF_WEB_EID_MOCK__?: unknown }
    win.__PF_WEB_EID_MOCK__ = {
      status: async () => ({ extension: '2.0.0', nativeApp: '2.0.0' }),
      authenticate: async () => {
        throw Object.assign(new Error('user_timeout'), { code: 'ERR_WEBEID_USER_TIMEOUT' })
      },
    }
    fetchMock.mockResolvedValueOnce({
      data: { nonce: 'n1', origin: 'http://localhost:3002' },
    })
    const { useEidPrefill } = await import('../../composables/useEidPrefill')
    const { readCard } = useEidPrefill()
    await expect(readCard()).rejects.toThrow('web_eid_timeout')
  })

  it('maps extension unavailable from status', async () => {
    // Non-loopback: status() hard-gate must fail fast (skipped on localhost).
    const win = window as Window & { __PF_WEB_EID_MOCK__?: unknown; location: { origin: string } }
    win.location = { origin: 'https://staging.petsfollow.app' }
    win.__PF_WEB_EID_MOCK__ = {
      status: async () => {
        throw Object.assign(new Error('Web-eID extension is not available'), {
          code: 'ERR_WEBEID_EXTENSION_UNAVAILABLE',
          name: 'ExtensionUnavailableError',
        })
      },
    }
    fetchMock.mockResolvedValueOnce({
      data: { nonce: 'n1', origin: 'https://staging.petsfollow.app' },
    })
    const { useEidPrefill } = await import('../../composables/useEidPrefill')
    const { readCard } = useEidPrefill()
    await expect(readCard()).rejects.toThrow('web_eid_extension_unavailable')
  })

  it('maps extension unavailable to loopback code on localhost', async () => {
    const win = window as Window & { __PF_WEB_EID_MOCK__?: unknown }
    win.__PF_WEB_EID_MOCK__ = {
      status: async () => ({ extension: '2.0.0', nativeApp: '2.0.0' }),
      authenticate: async () => {
        throw Object.assign(new Error('Web-eID extension is not available'), {
          code: 'ERR_WEBEID_EXTENSION_UNAVAILABLE',
          name: 'ExtensionUnavailableError',
        })
      },
    }
    fetchMock.mockResolvedValueOnce({
      data: { nonce: 'n1', origin: 'http://localhost:3002' },
    })
    const { useEidPrefill } = await import('../../composables/useEidPrefill')
    const { readCard } = useEidPrefill()
    await expect(readCard()).rejects.toThrow('web_eid_loopback_unavailable')
  })
})
