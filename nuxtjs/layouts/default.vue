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
const showNav = computed(() => {
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
  return !bare.includes(route.path) && !route.path.startsWith('/register')
})
const { fetchUser } = useProUser()
const requestsBadge = ref(0)

const navItems = computed<ProNavItem[]>(() => [
  { to: '/dashboard', label: t('nav.dashboard'), exact: true, icon: 'dashboard' },
  { to: '/clients', label: t('nav.clients'), icon: 'clients' },
  { to: '/requests', label: t('nav.requests'), icon: 'requests', badge: requestsBadge.value },
  { to: '/messages', label: t('nav.messages'), icon: 'messages' },
  { to: '/commissions', label: t('nav.commissions'), icon: 'payments' },
  { to: '/settings', label: t('nav.settings'), icon: 'settings' },
])

async function loadRequestsBadge() {
  if (!showNav.value) return
  try {
    const res: any = await $fetch('/api/vet/overview')
    const data = res.data ?? res ?? {}
    const links = Number(data.pendingLinkRequests ?? 0)
    const visits = Number(data.pendingVisits ?? 0)
    requestsBadge.value = links + visits
  } catch {
    requestsBadge.value = 0
  }
}

onMounted(async () => {
  await fetchUser()
  await loadRequestsBadge()
})

watch(() => route.path, (path) => {
  if (path === '/requests' || path === '/dashboard') {
    loadRequestsBadge()
  }
})
</script>
