import { isPublicFlagOn } from '~/utils/public-feature-flag'

export type PracticeSiteSummary = {
  id: string
  name: string
  isPrimary: boolean
  timezone: string
  active?: boolean
}

const SITE_ALL = 'all'
const STORAGE_KEY = 'pf_site_id'

export function usePracticeSites() {
  const { user } = useProUser()
  const runtimeConfig = useRuntimeConfig()
  const selectedSiteId = useState<string>('pf-selected-site-id', () => '')

  const sitesUiEnabled = computed(() => isPublicFlagOn(runtimeConfig.public.sitesUiEnabled))

  const sites = computed<PracticeSiteSummary[]>(() => {
    const raw = (user.value as any)?.sites
    if (!Array.isArray(raw)) return []
    return raw.map((s: any) => ({
      id: String(s.id || ''),
      name: String(s.name || ''),
      isPrimary: !!s.isPrimary,
      timezone: String(s.timezone || 'Europe/Brussels'),
      active: s.active !== false,
    })).filter((s: PracticeSiteSummary) => !!s.id)
  })

  const defaultSiteId = computed(() => {
    const fromTeam = String((user.value as any)?.defaultSiteId || '')
    if (fromTeam && sites.value.some((s) => s.id === fromTeam)) return fromTeam
    const primary = sites.value.find((s) => s.isPrimary)
    return primary?.id || sites.value[0]?.id || ''
  })

  const multiSite = computed(() => sites.value.length > 1)

  const effectiveSiteId = computed(() => {
    // Soft-GA off: always primary / team default (ignore switcher + storage multi).
    if (!sitesUiEnabled.value) return defaultSiteId.value
    if (!selectedSiteId.value) return defaultSiteId.value
    if (selectedSiteId.value === SITE_ALL) return SITE_ALL
    if (sites.value.some((s) => s.id === selectedSiteId.value)) return selectedSiteId.value
    return defaultSiteId.value
  })

  /** Query value for calendar/consult list (all | concrete id | empty for primary mono). */
  const calendarSiteQuery = computed(() => {
    // Soft-GA off: pin to primary/default — API empty siteId means ALL sites.
    if (!sitesUiEnabled.value) {
      return defaultSiteId.value || ''
    }
    const id = effectiveSiteId.value
    if (!id || id === defaultSiteId.value && !multiSite.value) return ''
    return id
  })

  /** True when switcher is on aggregated "all sites" view. */
  const isAggregatedView = computed(() => sitesUiEnabled.value && effectiveSiteId.value === SITE_ALL)

  /** Concrete site for create/edit (never "all"). */
  const concreteSiteId = computed(() => {
    const id = effectiveSiteId.value
    if (!id || id === SITE_ALL) return defaultSiteId.value
    return id
  })

  const concreteSiteName = computed(() => {
    const id = concreteSiteId.value
    return sites.value.find((s) => s.id === id)?.name || ''
  })

  function initFromStorage() {
    if (!import.meta.client) return
    if (!sitesUiEnabled.value) {
      selectedSiteId.value = defaultSiteId.value
      return
    }
    try {
      const stored = localStorage.getItem(STORAGE_KEY) || ''
      if (stored === SITE_ALL || sites.value.some((s) => s.id === stored)) {
        selectedSiteId.value = stored
        return
      }
    } catch { /* ignore */ }
    if (!selectedSiteId.value) {
      selectedSiteId.value = defaultSiteId.value
    }
  }

  function setSiteId(id: string) {
    selectedSiteId.value = id
    if (import.meta.client) {
      try {
        localStorage.setItem(STORAGE_KEY, id)
      } catch { /* ignore */ }
    }
  }

  function withSiteQuery(url: string, siteId?: string): string {
    const sid = siteId === undefined ? calendarSiteQuery.value : siteId
    if (!sid) return url
    const sep = url.includes('?') ? '&' : '?'
    return `${url}${sep}siteId=${encodeURIComponent(sid)}`
  }

  return {
    SITE_ALL,
    sites,
    sitesUiEnabled,
    multiSite,
    selectedSiteId,
    effectiveSiteId,
    calendarSiteQuery,
    concreteSiteId,
    concreteSiteName,
    isAggregatedView,
    defaultSiteId,
    initFromStorage,
    setSiteId,
    withSiteQuery,
  }
}
