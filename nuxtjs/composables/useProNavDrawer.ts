const MOBILE_NAV_MQ = '(max-width: 960px)'

type NavDrawerGlobals = typeof globalThis & { __pfProNavMqBound?: boolean }

/** Mobile nav drawer open state — shared by ProTopbar (toggle) and ProSidebar (panel). */
export function useProNavDrawer() {
  const open = useState('pf-pro-nav-drawer-open', () => false)
  const isMobileNav = useState('pf-pro-nav-drawer-mobile', () => false)

  function openDrawer() {
    if (!isMobileNav.value) return
    open.value = true
  }

  function closeDrawer() {
    open.value = false
  }

  function toggleDrawer() {
    if (!isMobileNav.value) {
      open.value = false
      return
    }
    open.value = !open.value
  }

  if (import.meta.client) {
    const g = globalThis as NavDrawerGlobals
    if (!g.__pfProNavMqBound) {
      g.__pfProNavMqBound = true
      const mq = window.matchMedia(MOBILE_NAV_MQ)
      const sync = () => {
        isMobileNav.value = mq.matches
        if (!mq.matches) open.value = false
      }
      sync()
      mq.addEventListener('change', sync)
    }
  }

  return {
    open,
    isMobileNav,
    openDrawer,
    closeDrawer,
    toggleDrawer,
  }
}
