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

  it('uploadViewerFile posts multipart to import', async () => {
    fetchMock.mockResolvedValueOnce({
      data: { firstname: 'A', lastname: 'B', niss: '96072399886' },
    })
    const { useEidPrefill } = await import('../../composables/useEidPrefill')
    const { uploadViewerFile } = useEidPrefill()
    const file = new File(['<x/>'], 'x.eid', { type: 'application/xml' })
    const identity = await uploadViewerFile(file)
    expect(identity.lastname).toBe('B')
    expect(fetchMock).toHaveBeenCalledWith('/api/vet/eid/import', {
      method: 'POST',
      body: expect.any(FormData),
    })
  })
})
