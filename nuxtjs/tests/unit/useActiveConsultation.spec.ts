import { beforeEach, describe, expect, it, vi } from 'vitest'

const store = new Map<string, string>()
const stateStore = new Map<string, { value: unknown }>()
const navigateTo = vi.fn()

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

vi.stubGlobal('navigateTo', navigateTo)

describe('useActiveConsultation resume isolation', () => {
  beforeEach(() => {
    store.clear()
    stateStore.clear()
    navigateTo.mockReset()
    vi.resetModules()
    vi.stubGlobal('navigateTo', navigateTo)
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
  })

  it('peek keeps token; consume deletes it; no cross-user leak', async () => {
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    a.saveResumeForEmail('vet.demo@petsfollow.test', {
      visitId: 'visit-a',
      clientId: 'client-a',
      petId: 'pet-a',
      path: '/clients',
      updatedAt: Date.now(),
    })
    expect(a.peekResumeForEmail('vet.colleague@petsfollow.test')).toBeNull()
    const peeked = a.peekResumeForEmail('vet.demo@petsfollow.test')
    expect(peeked?.visitId).toBe('visit-a')
    expect(a.peekResumeForEmail('vet.demo@petsfollow.test')?.visitId).toBe('visit-a')
    const token = a.consumeResumeForEmail('vet.demo@petsfollow.test')
    expect(token?.visitId).toBe('visit-a')
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
    expect(a.peekResumeForEmail('vet.demo@petsfollow.test')).toBeNull()
    expect(store.get('pf_consult_resume')).toBe('{}')
  })

  it('flushBeforeSuspend sets suspendDiscard and writes resume only then', async () => {
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    a.open.value = true
    a.clientId.value = 'client-a'
    a.resumeVisitId.value = 'visit-a'
    a.resumePetId.value = 'pet-a'
    await a.autosaveOnly()
    expect(store.get('pf_consult_resume')).toBeUndefined()
    await a.flushBeforeSuspend('vet.demo@petsfollow.test', '/clients')
    expect(a.suspendDiscard.value).toBe(true)
    expect(a.open.value).toBe(false)
    expect(a.clientId.value).toBe('')
    const token = JSON.parse(store.get('pf_consult_resume') || '{}')
    expect(token['vet.demo@petsfollow.test']?.visitId).toBe('visit-a')
  })

  it('openForVisit opens modal shell with keepVisit (no page navigate)', async () => {
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    a.openForVisit({ visitId: 'visit-rdv', clientId: 'client-a', petId: 'pet-a' })
    expect(a.open.value).toBe(true)
    expect(a.resumeVisitId.value).toBe('visit-rdv')
    expect(a.resumeKeepVisit.value).toBe(true)
    expect(navigateTo).not.toHaveBeenCalled()
    await a.flushBeforeSuspend('vet.demo@petsfollow.test', '/calendar')
    expect(a.resumeKeepVisit.value).toBe(false)
    const token = a.peekResumeForEmail('vet.demo@petsfollow.test')
    expect(token?.keepVisit).toBe(true)
    a.openResume(token!)
    expect(a.open.value).toBe(true)
    expect(a.resumeKeepVisit.value).toBe(true)
    a.close()
    a.openForClient('client-b')
    expect(a.open.value).toBe(true)
    expect(a.resumeKeepVisit.value).toBe(false)
  })

  it('tryResumeForCurrentUser opens modal once and clears token', async () => {
    const { useActiveConsultation } = await import('../../composables/useActiveConsultation')
    const a = useActiveConsultation()
    a.saveResumeForEmail('vet.demo@petsfollow.test', {
      visitId: 'visit-b',
      clientId: 'client-b',
      petId: 'pet-b',
      path: '/clients',
      updatedAt: Date.now(),
    })
    expect(a.tryResumeForCurrentUser('vet.demo@petsfollow.test')).toBe(true)
    expect(a.open.value).toBe(true)
    expect(a.resumeVisitId.value).toBe('visit-b')
    expect(navigateTo).not.toHaveBeenCalled()
    expect(a.peekResumeForEmail('vet.demo@petsfollow.test')).toBeNull()
    expect(a.tryResumeForCurrentUser('vet.demo@petsfollow.test')).toBe(false)
  })
})
