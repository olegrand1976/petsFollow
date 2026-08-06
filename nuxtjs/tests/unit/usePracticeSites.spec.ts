import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'

const stateStore = new Map<string, { value: unknown }>()

const userRef = ref({
  sites: [
    { id: 'site-a', name: 'Primary', isPrimary: true, timezone: 'Europe/Brussels' },
    { id: 'site-b', name: 'Antenne', isPrimary: false, timezone: 'Europe/Brussels' },
  ],
  defaultSiteId: 'site-a',
})

vi.stubGlobal('computed', computed)
vi.stubGlobal('ref', ref)
vi.stubGlobal('useState', (key: string, init?: () => unknown) => {
  if (!stateStore.has(key)) {
    stateStore.set(key, ref(init ? init() : null))
  }
  return stateStore.get(key)!
})
vi.stubGlobal('useRuntimeConfig', () => ({ public: { sitesUiEnabled: true } }))
vi.stubGlobal('useProUser', () => ({ user: userRef }))

describe('usePracticeSites', () => {
  beforeEach(() => {
    stateStore.clear()
    vi.resetModules()
    vi.stubGlobal('computed', computed)
    vi.stubGlobal('ref', ref)
    vi.stubGlobal('useState', (key: string, init?: () => unknown) => {
      if (!stateStore.has(key)) {
        stateStore.set(key, ref(init ? init() : null))
      }
      return stateStore.get(key)!
    })
    vi.stubGlobal('useRuntimeConfig', () => ({ public: { sitesUiEnabled: true } }))
    vi.stubGlobal('useProUser', () => ({ user: userRef }))
  })

  it('withSiteQuery appends siteId and exposes sitesUiEnabled', async () => {
    const { usePracticeSites } = await import('../../composables/usePracticeSites')
    const { withSiteQuery, setSiteId, SITE_ALL, sitesUiEnabled, multiSite, calendarSiteQuery } = usePracticeSites()
    expect(sitesUiEnabled.value).toBe(true)
    expect(multiSite.value).toBe(true)

    setSiteId('site-b')
    expect(calendarSiteQuery.value).toBe('site-b')
    expect(withSiteQuery('/api/vet/consultations')).toContain('siteId=site-b')
    expect(withSiteQuery('/api/vet/consultations?q=x')).toContain('&siteId=site-b')
    expect(withSiteQuery('/api/vet/calendar', SITE_ALL)).toContain('siteId=all')
  })
})
