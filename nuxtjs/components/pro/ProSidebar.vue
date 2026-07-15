<template>
  <aside class="pro-sidebar">
    <nav>
      <NuxtLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        :exact="item.exact"
        class="pro-sidebar__link"
      >
        <span class="pro-sidebar__icon" aria-hidden="true" v-html="iconSvg(item.icon)" />
        <span>{{ item.label }}</span>
      </NuxtLink>
    </nav>
  </aside>
</template>

<script setup lang="ts">
export type ProNavIcon = 'dashboard' | 'clients' | 'messages' | 'settings' | 'admin' | 'users' | 'payments'

export type ProNavItem = {
  to: string
  label: string
  exact?: boolean
  icon: ProNavIcon
}

defineProps<{
  items: ProNavItem[]
}>()

const icons: Record<ProNavIcon, string> = {
  dashboard: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><rect x="3" y="3" width="8" height="8" rx="2" stroke="currentColor" stroke-width="1.75"/><rect x="13" y="3" width="8" height="5" rx="2" stroke="currentColor" stroke-width="1.75"/><rect x="13" y="10" width="8" height="11" rx="2" stroke="currentColor" stroke-width="1.75"/><rect x="3" y="13" width="8" height="8" rx="2" stroke="currentColor" stroke-width="1.75"/></svg>',
  clients: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><circle cx="9" cy="8" r="3" stroke="currentColor" stroke-width="1.75"/><path d="M3 20c0-3.3 2.7-6 6-6s6 2.7 6 6" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/><circle cx="17" cy="9" r="2.5" stroke="currentColor" stroke-width="1.75"/><path d="M15 20c.3-2.2 1.8-4 4-4" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/></svg>',
  messages: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M4 6h16v10H7l-3 3V6z" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/></svg>',
  settings: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.75"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/></svg>',
  admin: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M4 20V10l8-6 8 6v10H4z" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round"/><path d="M9 20v-6h6v6" stroke="currentColor" stroke-width="1.75"/></svg>',
  users: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="8" r="4" stroke="currentColor" stroke-width="1.75"/><path d="M4 21c0-4 3.6-7 8-7s8 3 8 7" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/></svg>',
  payments: '<svg width="18" height="18" viewBox="0 0 24 24" fill="none"><rect x="3" y="6" width="18" height="12" rx="2" stroke="currentColor" stroke-width="1.75"/><path d="M3 10h18" stroke="currentColor" stroke-width="1.75"/></svg>',
}

function iconSvg(name: ProNavIcon) {
  return icons[name] ?? icons.dashboard
}
</script>

<style scoped>
.pro-sidebar__link {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.pro-sidebar__icon {
  display: inline-flex;
  opacity: 0.9;
}
</style>
