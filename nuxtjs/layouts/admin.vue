<template>
  <div class="pro-app">
    <ProSidebar :items="navItems" />
    <div class="pro-app-body">
      <ProTopbar home-link="/admin" :show-notifications="false" />
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

const { fetchUser } = useProUser()

const navItems: ProNavItem[] = [
  { to: '/admin', label: 'Tableau de bord', exact: true, icon: 'admin' },
  { to: '/admin/users', label: 'Inscriptions', icon: 'users' },
  { to: '/admin/payments', label: 'Paiements', icon: 'payments' },
]

onMounted(() => fetchUser())
</script>
