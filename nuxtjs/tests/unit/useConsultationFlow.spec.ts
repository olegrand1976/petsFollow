import { beforeEach, describe, expect, it, vi } from 'vitest'

/**
 * discardIfUnsaved : le nettoyage d'orphelin walk-in ne doit JAMAIS annuler
 * un RDV agenda existant (preserveVisit — consultation ouverte via openForVisit).
 */

const fetchMock = vi.fn()

vi.stubGlobal('$fetch', fetchMock)
vi.stubGlobal('ref', <T>(v: T) => ({ value: v }))
vi.stubGlobal('computed', <T>(fn: () => T) => ({
  get value() {
    return fn()
  },
}))
vi.stubGlobal('useRuntimeConfig', () => ({ public: {} }))

describe('useConsultationFlow discardIfUnsaved', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    fetchMock.mockResolvedValue({})
  })

  it('cancels an unsaved walk-in visit (orphan cleanup)', async () => {
    const { useConsultationFlow } = await import('../../composables/useConsultationFlow')
    const flow = useConsultationFlow()
    flow.visitId.value = 'visit-walkin'
    await flow.discardIfUnsaved()
    expect(fetchMock).toHaveBeenCalledWith('/api/visits/visit-walkin', {
      method: 'PATCH',
      body: { status: 'cancelled' },
    })
  })

  it('never cancels an existing agenda RDV (preserveVisit)', async () => {
    const { useConsultationFlow } = await import('../../composables/useConsultationFlow')
    const flow = useConsultationFlow()
    flow.visitId.value = 'visit-rdv'
    flow.preserveVisit.value = true
    await flow.discardIfUnsaved()
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('skips cancel once the report is saved', async () => {
    const { useConsultationFlow } = await import('../../composables/useConsultationFlow')
    const flow = useConsultationFlow()
    flow.visitId.value = 'visit-saved'
    flow.afterSaved()
    await flow.discardIfUnsaved()
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('reset clears preserveVisit', async () => {
    const { useConsultationFlow } = await import('../../composables/useConsultationFlow')
    const flow = useConsultationFlow()
    flow.preserveVisit.value = true
    flow.reset()
    expect(flow.preserveVisit.value).toBe(false)
  })
})
