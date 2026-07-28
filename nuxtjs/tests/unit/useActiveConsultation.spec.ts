import { beforeEach, describe, expect, it, vi } from 'vitest'

const store = new Map<string, string>()
const stateStore = new Map<string, { value: unknown }>()

vi.stubGlobal('localStorage', {
  getItem: (k: string) => store.get(k) ?? null,
  setItem: (k: string, v: string) => { store.set(k, v) },
  removeItem: (k: string) => { store.delete(k) },
  clear: () => { store.clear() },
})

vi.stubGlobal('useState', (key: string, init?: () => unknown) => {
  if (!stateStore.has(key)) {
    stateStore.set(key, { value: init ? init() : null })
  }
  return stateStore.get(key)!
})

describe('useActiveConsultation resume isolation', () => {
  beforeEach(() => {
    store.clear()
    stateStore.clear()
    vi.resetModules()
  })

  it('stores and consumes resume per email without cross-user leak', async () => {
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    a.saveResumeForEmail('vet.demo@petsfollow.test', {
      visitId: 'visit-a',
      clientId: 'client-a',
      petId: 'pet-a',
      path: '/clients',
      updatedAt: Date.now(),
    })
    expect(a.consumeResumeForEmail('vet.colleague@petsfollow.test')).toBeNull()
    const token = a.consumeResumeForEmail('vet.demo@petsfollow.test')
    expect(token?.visitId).toBe('visit-a')
    expect(token?.clientId).toBe('client-a')
    // True consume: second read is empty.
    expect(a.consumeResumeForEmail('vet.demo@petsfollow.test')).toBeNull()
  })

  it('rejects expired resume tokens', async () => {
    store.set('pf_consult_resume', JSON.stringify({
      'vet.demo@petsfollow.test': {
        visitId: 'visit-old',
        clientId: 'client-a',
        petId: 'pet-a',
        path: '/clients',
        updatedAt: Date.now() - 13 * 60 * 60 * 1000,
      },
    }))
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    expect(a.consumeResumeForEmail('vet.demo@petsfollow.test')).toBeNull()
    // Expired entry removed from storage.
    expect(store.get('pf_consult_resume')).toBe('{}')
  })

  it('flushBeforeSuspend sets suspendDiscard even without snapshot when open', async () => {
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    a.open.value = true
    a.clientId.value = 'client-a'
    a.resumeVisitId.value = 'visit-a'
    a.resumePetId.value = 'pet-a'
    await a.flushBeforeSuspend('vet.demo@petsfollow.test', '/clients')
    expect(a.suspendDiscard.value).toBe(true)
    expect(a.open.value).toBe(false)
    expect(a.clientId.value).toBe('')
    const token = JSON.parse(store.get('pf_consult_resume') || '{}')
    expect(token['vet.demo@petsfollow.test']?.visitId).toBe('visit-a')
  })

  it('openForClient sets preferred pet without visit resume', async () => {
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    a.resumeVisitId.value = 'stale-visit'
    a.openForClient('client-b', 'pet-b')
    expect(a.open.value).toBe(true)
    expect(a.clientId.value).toBe('client-b')
    expect(a.resumePetId.value).toBe('pet-b')
    expect(a.resumeVisitId.value).toBe('')
  })

  it('openForClient without preferred pet clears resumePetId', async () => {
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    a.resumePetId.value = 'pet-old'
    a.openForClient('client-c')
    expect(a.clientId.value).toBe('client-c')
    expect(a.resumePetId.value).toBe('')
    expect(a.resumeVisitId.value).toBe('')
  })
})
