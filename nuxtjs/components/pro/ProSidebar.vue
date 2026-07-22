<template>
  <aside class="pro-sidebar">
    <nav>
      <NuxtLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        :exact="item.exact"
        class="pro-sidebar__link"
        :data-testid="navTestId(item.to)"
      >
        <span class="pro-sidebar__icon" aria-hidden="true">
          <ProIcon :name="iconName(item.icon)" :size="18" />
        </span>
        <span class="pro-sidebar__label">{{ item.label }}</span>
        <ProBadge
          v-if="item.badge && item.badge > 0"
          variant="danger"
          class="pro-sidebar__badge"
          :data-testid="badgeTestId(item.to)"
        >
          {{ item.badge > 99 ? '99+' : item.badge }}
        </ProBadge>
      </NuxtLink>
    </nav>
  </aside>
</template>

<script setup lang="ts">
export type ProNavIcon =
  | 'dashboard'
  | 'clients'
  | 'pets'
  | 'messages'
  | 'settings'
  | 'admin'
  | 'users'
  | 'payments'
  | 'requests'
  | 'calendar'
  | 'recommend'
  | 'description'

export type ProNavItem = {
  to: string
  label: string
  exact?: boolean
  icon: ProNavIcon
  badge?: number
}

defineProps<{
  items: ProNavItem[]
}>()

const icons: Record<ProNavIcon, string> = {
  dashboard: 'dashboard',
  clients: 'group',
  pets: 'pets',
  messages: 'chat',
  settings: 'settings',
  admin: 'admin_panel_settings',
  users: 'person',
  payments: 'payments',
  requests: 'inbox',
  calendar: 'calendar_month',
  recommend: 'handshake',
  description: 'description',
}

function iconName(name: ProNavIcon) {
  return icons[name] ?? icons.dashboard
}

function navTestId(to: string) {
  if (to === '/calendar') return 'nav-calendar'
  if (to === '/clients') return 'nav-clients'
  if (to === '/pets') return 'nav-pets'
  if (to === '/messages') return 'nav-messages'
  if (to === '/requests') return 'nav-requests'
  return undefined
}

function badgeTestId(to: string) {
  if (to === '/calendar') return 'nav-calendar-badge'
  if (to === '/clients') return 'nav-clients-badge'
  if (to === '/messages') return 'nav-messages-badge'
  if (to === '/requests') return 'nav-requests-badge'
  return undefined
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

.pro-sidebar__label {
  flex: 1;
  min-width: 0;
}

.pro-sidebar__badge {
  margin-left: auto;
  flex-shrink: 0;
  min-width: 1.25rem;
  justify-content: center;
  padding: 0.15rem 0.4rem;
  background: var(--pf-vet-alert);
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: none;
  line-height: 1.2;
}
</style>
