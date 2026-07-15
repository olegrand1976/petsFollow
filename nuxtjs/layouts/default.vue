<template>
  <div class="pro-app">
    <ProSidebar v-if="showNav" :items="navItems" />
    <div class="pro-app-body">
      <ProTopbar v-if="showNav" home-link="/" settings-link="/settings" />
      <main class="pro-main main">
        <div class="pro-main-inner">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProNavItem } from '~/components/pro/ProSidebar.vue'

const route = useRoute()
const showNav = computed(() => route.path !== '/login')
const { fetchUser } = useProUser()

const navItems: ProNavItem[] = [
  { to: '/', label: 'Dashboard', exact: true, icon: 'dashboard' },
  { to: '/clients', label: 'Clients', icon: 'clients' },
  { to: '/messages', label: 'Messagerie', icon: 'messages' },
  { to: '/settings', label: 'Paramètres', icon: 'settings' },
]

onMounted(() => fetchUser())
</script>
