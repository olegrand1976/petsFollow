<template>
  <div class="pro-app">
    <ProTopbar v-if="showNav" home-link="/dashboard" settings-link="/settings" />
    <div class="pro-app-shell">
      <ProSidebar v-if="showNav" :items="navItems" />
      <div class="pro-app-body">
        <main class="pro-main main">
          <div class="pro-main-inner">
            <slot />
          </div>
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProNavItem } from '~/components/pro/ProSidebar.vue'

const route = useRoute()
const { t } = useI18n()
const { user, fetchUser } = useProUser()
// SSR : résoudre /me avant rendu pour aligner hydratation (pas de shell fantôme).
if (!user.value) {
  await fetchUser().catch(() => null)
}
const showNav = computed(() => {
  // Pas de shell Pro tant que /api/me n'a pas confirmé le rôle (évite login sous topbar).
  if (!user.value?.role) return false
  const bare = [
    '/login',
    '/register',
    '/register/sent',
    '/confirm-email',
    '/forgot-password',
    '/reset-password',
    '/welcome',
    '/',
  ]
  return !bare.includes(route.path)
    && !route.path.startsWith('/register')
    && !route.path.startsWith('/invite')
})
const { count: messagesBadge } = useProNotifications()
const { clientsBadge, calendarBadge, petsBadge, refresh: refreshNavBadges } = useNavBadges()

const navItems = computed<ProNavItem[]>(() => {
  const items: ProNavItem[] = [
    { to: '/dashboard', label: t('nav.dashboard'), exact: true, icon: 'dashboard' },
    { to: '/clients', label: t('nav.clients'), icon: 'clients', badge: clientsBadge.value },
    { to: '/pets', label: t('nav.pets'), icon: 'pets', badge: petsBadge.value },
    { to: '/calendar', label: t('nav.calendar'), icon: 'calendar', badge: calendarBadge.value },
    { to: '/messages', label: t('nav.messages'), icon: 'messages', badge: messagesBadge.value },
    { to: '/produits', label: t('nav.products'), icon: 'description' },
  ]
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
  if (!user.value) {
    try {
      await fetchUser()
    } catch { /* 401 handled by middleware */ }
  }
  await loadNavBadges()
})

watch(() => route.path, (path) => {
  if (path === '/calendar' || path === '/clients' || path === '/dashboard' || path === '/pets') {
    loadNavBadges()
  }
})
</script>
