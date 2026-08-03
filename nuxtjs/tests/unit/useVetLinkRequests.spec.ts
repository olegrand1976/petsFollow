import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'

vi.stubGlobal('ref', ref)
vi.stubGlobal('useApiError', () => ({ mapError: (e: any) => e?.message || 'error' }))

type Handler = (url: string, opts?: any) => any
let fetchImpl: Handler
vi.stubGlobal('$fetch', (url: string, opts?: any) => Promise.resolve(fetchImpl(url, opts)))

const req = (id: string) => ({ id, clientName: `Client ${id}`, clientEmail: `${id}@petsfollow.test` })

async function load() {
  const { useVetLinkRequests } = await import('../../composables/useVetLinkRequests')
  return useVetLinkRequests({ canManage: () => true })
}

describe('useVetLinkRequests', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  it('ferme la modale quand la dernière invitation est acceptée', async () => {
    let pending = [req('a'), req('b')]
    fetchImpl = (url) => {
      if (url.endsWith('/accept')) {
        pending = pending.filter((r) => !url.includes(r.id))
        return {}
      }
      return { data: pending }
    }
    const inv = await load()
    await inv.load()
    inv.open.value = true
    expect(inv.items.value).toHaveLength(2)

    await inv.accept('a')
    expect(inv.items.value.map((r) => r.id)).toEqual(['b'])
    expect(inv.open.value).toBe(true)

    await inv.accept('b')
    expect(inv.items.value).toHaveLength(0)
    expect(inv.open.value).toBe(false)
  })

  it('ferme la modale quand la dernière invitation est refusée', async () => {
    let pending = [req('a')]
    fetchImpl = (url) => {
      if (url.endsWith('/reject')) {
        pending = []
        return {}
      }
      return { data: pending }
    }
    const inv = await load()
    await inv.load()
    inv.open.value = true

    await inv.reject('a')
    expect(inv.open.value).toBe(false)
  })

  it('liste vide renvoyée par l’API (data null) → items est un tableau', async () => {
    fetchImpl = () => ({ data: null })
    const inv = await load()
    await inv.load()
    expect(inv.items.value).toEqual([])
  })

  it('garde la modale ouverte si l’action échoue', async () => {
    fetchImpl = (url) => {
      if (url.endsWith('/accept')) throw new Error('boom')
      return { data: [req('a')] }
    }
    const inv = await load()
    await inv.load()
    inv.open.value = true

    await inv.accept('a')
    expect(inv.open.value).toBe(true)
    expect(inv.error.value).toBe('boom')
    expect(inv.busyId.value).toBe('')
  })

  it('garde la modale ouverte si le rechargement échoue (erreur affichée)', async () => {
    fetchImpl = (url) => {
      if (url.endsWith('/accept')) return {}
      throw new Error('reload down')
    }
    const inv = await load()
    inv.open.value = true

    await inv.accept('a')
    expect(inv.items.value).toEqual([])
    expect(inv.error.value).toBe('reload down')
    expect(inv.open.value).toBe(true)
  })

  it('sans permission shares.manage, aucune invitation chargée', async () => {
    fetchImpl = () => {
      throw new Error('should not be called')
    }
    const { useVetLinkRequests } = await import('../../composables/useVetLinkRequests')
    const inv = useVetLinkRequests({ canManage: () => false })
    await inv.load()
    expect(inv.items.value).toEqual([])
    expect(inv.error.value).toBe('')
  })
})
