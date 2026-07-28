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
            <template v-if="!deskUiBlocked">
              <slot />
              <ProActiveConsultationHost />
            </template>
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
const { canPractice } = usePracticePerms()
const runtimeConfig = useRuntimeConfig()
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

const navItems = computed<ProNavItem[]>(() => {
  const pharmacyOn = Boolean(runtimeConfig.public.pharmacyEnabled)
  const prescriptionsOn = Boolean(runtimeConfig.public.prescriptionsEnabled)
  const billitOn = Boolean(runtimeConfig.public.billitEnabled)
  const day = t('nav.section.day')
  const patients = t('nav.section.patients')
  const clinic = t('nav.section.clinic')
  const finance = t('nav.section.finance')
  const offer = t('nav.section.offer')
  const practice = t('nav.section.practice')
  const tagDev = t('nav.tagDev')

  const items: ProNavItem[] = [
    { to: '/dashboard', label: t('nav.dashboard'), exact: true, icon: 'dashboard', section: day },
  ]
  if (canPractice('calendar.manage')) {
    items.push({ to: '/calendar', label: t('nav.calendar'), icon: 'calendar', badge: calendarBadge.value, section: day })
    items.push({ to: '/consultations', label: t('nav.consultations'), icon: 'clinical_notes', section: day })
  }
  if (canPractice('messaging')) {
    items.push({ to: '/messages', label: t('nav.messages'), icon: 'messages', badge: messagesBadge.value, section: day })
  }

  if (canPractice('clients.read')) {
    items.push({ to: '/clients', label: t('nav.clients'), icon: 'clients', badge: clientsBadge.value, section: patients })
  }
  if (canPractice('pets.read')) {
    items.push({ to: '/pets', label: t('nav.pets'), icon: 'pets', badge: petsBadge.value, section: patients })
  }

  if (prescriptionsOn && canPractice('pets.read')) {
    items.push({ to: '/ordonnances', label: t('nav.prescriptions'), icon: 'clinical_notes', tag: tagDev, section: clinic })
  }
  if (pharmacyOn && canPractice('pharmacy.read')) {
    items.push(
      { to: '/medicaments', label: t('nav.medicaments'), icon: 'medication', tag: tagDev, section: clinic },
      { to: '/stock', label: t('nav.stock'), icon: 'inventory_2', tag: tagDev, section: clinic },
      { to: '/daf', label: t('nav.daf'), icon: 'local_shipping', tag: tagDev, section: clinic },
    )
  }

  if (billitOn && (canPractice('clients.write') || canPractice('practice.settings'))) {
    items.push({ to: '/invoicing', label: t('nav.invoicing'), icon: 'receipt', tag: tagDev, section: finance })
  }
  if (canPractice('commissions.view')) {
    items.push({ to: '/commissions', label: t('nav.commissions'), icon: 'payments', section: finance })
  }

  items.push({ to: '/produits', label: t('nav.products'), icon: 'description', section: offer })
  if (canPractice('clients.write')) {
    items.push({ to: '/recommend', label: t('nav.recommend'), icon: 'recommend', section: offer })
  }

  items.push({ to: '/team', label: t('nav.team'), icon: 'groups', section: practice })
  items.push({ to: '/settings', label: t('nav.settings'), icon: 'settings', section: practice })

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
    // Flag stale + session active ⇒ bootstrap a levé la veille : continuer l'init.
    if (desk.locked.value || isDeskLockedFlag()) {
      stopPolling()
      return
    }
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
  if (path === '/calendar' || path === '/consultations' || path === '/clients' || path === '/dashboard' || path === '/pets') {
    loadNavBadges()
  }
})
</script>
