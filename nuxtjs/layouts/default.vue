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

const navItems = computed<ProNavItem[]>(() => [
  { to: '/dashboard', label: t('nav.dashboard'), exact: true, icon: 'dashboard' },
  { to: '/clients', label: t('nav.clients'), icon: 'clients' },
  { to: '/messages', label: t('nav.messages'), icon: 'messages' },
  { to: '/commissions', label: t('nav.commissions'), icon: 'payments' },
  { to: '/settings', label: t('nav.settings'), icon: 'settings' },
])

onMounted(() => fetchUser())
</script>
