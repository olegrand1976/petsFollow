<template>
  <div class="pro-app">
    <ProTopbar v-if="showNav" home-link="/dashboard" settings-link="/settings" />
    <ProDeskLockOverlay />
    <div class="pro-app-shell">
      <ProSidebar v-if="showNav" :items="navItems" />
      <div class="pro-app-body">
        <main class="pro-main main">
          <div class="pro-main-inner">
            <!-- Lock / switch : démonter la page pour ne pas laisser de PHI dans le DOM. -->
            <slot v-if="!deskUiBlocked" />
          </div>
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProNavItem } from '~/components/pro/ProSidebar.vue'
import { isPracticeStaffRole } from '~/composables/useAuth'
import { isDeskLockedFlag } from '~/composables/useDeskSession'

const route = useRoute()
const { t } = useI18n()
const { user, fetchUser } = useProUser()
const desk = useDeskSession()
const deskLocked = computed(() => desk.locked.value)
const deskUiBlocked = computed(() => desk.uiBlocked.value)
const {
  count: messagesBadge,
  stopPolling,
} = useProNotifications()
const { clientsBadge, calendarBadge, petsBadge, refresh: refreshNavBadges } = useNavBadges()

const bareShellPaths = new Set([
  '/login',
  '/register',
  '/register/sent',
  '/confirm-email',
  '/forgot-password',
  '/reset-password',
  '/welcome',
  '/',
])
function isBareShellPath(path: string) {
  return bareShellPaths.has(path)
    || path.startsWith('/register')
    || path.startsWith('/invite')
}

// SSR : /me avant rendu sur les pages shell uniquement (évite 401 bruyants sur landing/auth).
if (!user.value && !isBareShellPath(route.path) && !isDeskLockedFlag()) {
  await fetchUser().catch(() => null)
}
const showNav = computed(() => {
  // Pas de shell Pro tant que /api/me n'a pas confirmé le rôle (évite login sous topbar).
  // Veille / switch : pas de nav (overlay couvre — PHI + identité hors écran).
  if (deskUiBlocked.value) return false
  if (!user.value?.role) return false
  return !isBareShellPath(route.path)
})

const runtimeConfig = useRuntimeConfig()

const navItems = computed<ProNavItem[]>(() => {
  const prescriptionsOn = Boolean(runtimeConfig.public.prescriptionsEnabled)
  const items: ProNavItem[] = [
    { to: '/dashboard', label: t('nav.dashboard'), exact: true, icon: 'dashboard' },
    { to: '/clients', label: t('nav.clients'), icon: 'clients', badge: clientsBadge.value },
    { to: '/pets', label: t('nav.pets'), icon: 'pets', badge: petsBadge.value },
    { to: '/calendar', label: t('nav.calendar'), icon: 'calendar', badge: calendarBadge.value },
    { to: '/messages', label: t('nav.messages'), icon: 'messages', badge: messagesBadge.value },
  ]
  if (prescriptionsOn) {
    items.push({ to: '/ordonnances', label: t('nav.prescriptions'), icon: 'medication', tag: t('nav.tagDev') })
  }
  items.push(
    { to: '/invoicing', label: t('nav.invoicing'), icon: 'receipt', tag: t('nav.tagDev') },
    { to: '/produits', label: t('nav.products'), icon: 'description' },
  )
  if (user.value?.isReferenceVet === true) {
    items.push({ to: '/commissions', label: t('nav.commissions'), icon: 'payments' })
  }
  items.push(
    { to: '/team', label: t('nav.team'), icon: 'groups' },
    { to: '/recommend', label: t('nav.recommend'), icon: 'recommend' },
    { to: '/settings', label: t('nav.settings'), icon: 'settings' },
  )
  return items
})

async function loadNavBadges() {
  if (!showNav.value) return
  await refreshNavBadges()
}

onMounted(async () => {
  // Owner du cycle desk : layout (survit au démontage topbar en veille).
  if (isDeskLockedFlag() || deskLocked.value) {
    await desk.bootstrap()
    stopPolling()
    return
  }
  if (!user.value) {
    try {
      await fetchUser()
    } catch { /* 401 handled by middleware */ }
  }
  if (isPracticeStaffRole(user.value?.role)) {
    await desk.bootstrap()
  }
  await loadNavBadges()
})

onUnmounted(() => {
  desk.stopIdleWatch()
})

watch(
  () => desk.uiBlocked.value,
  (blocked) => {
    if (blocked) stopPolling()
  },
)

watch(() => route.fullPath, () => {
  desk.rememberCurrentPath()
})

watch(() => route.path, (path) => {
  if (path === '/calendar' || path === '/clients' || path === '/dashboard' || path === '/pets') {
    loadNavBadges()
  }
})
</script>
