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
        <span class="pro-sidebar__icon" aria-hidden="true">
          <ProIcon :name="iconName(item.icon)" :size="18" />
        </span>
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
  dashboard: 'dashboard',
  clients: 'group',
  messages: 'chat',
  settings: 'settings',
  admin: 'admin_panel_settings',
  users: 'person',
  payments: 'payments',
}

function iconName(name: ProNavIcon) {
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
