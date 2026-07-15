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

const { t } = useI18n()
const { fetchUser } = useProUser()

const navItems = computed<ProNavItem[]>(() => [
  { to: '/admin', label: t('nav.adminDashboard'), exact: true, icon: 'admin' },
  { to: '/admin/users', label: t('nav.adminUsers'), icon: 'users' },
  { to: '/admin/payments', label: t('nav.adminPayments'), icon: 'payments' },
])

onMounted(() => fetchUser())
</script>
