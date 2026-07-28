<template>
  <aside class="pro-sidebar">
    <nav>
      <template v-for="(entry, idx) in entries" :key="entry.key">
        <p
          v-if="entry.sectionLabel"
          class="pro-sidebar__section"
          :class="{ 'pro-sidebar__section--spaced': idx > 0 }"
        >
          {{ entry.sectionLabel }}
        </p>
        <NuxtLink
          :to="entry.item.to"
          :exact="entry.item.exact"
          class="pro-sidebar__link"
          :data-testid="navTestId(entry.item.to)"
        >
          <span class="pro-sidebar__icon" aria-hidden="true">
            <ProIcon :name="iconName(entry.item.icon)" :size="18" />
          </span>
          <span class="pro-sidebar__label">{{ entry.item.label }}</span>
          <ProBadge
            v-if="entry.item.tag"
            variant="warning"
            class="pro-sidebar__tag"
            :data-testid="tagTestId(entry.item.to)"
          >
            {{ entry.item.tag }}
          </ProBadge>
          <ProBadge
            v-else-if="entry.item.badge && entry.item.badge > 0"
            variant="danger"
            class="pro-sidebar__badge"
            :data-testid="badgeTestId(entry.item.to)"
          >
            {{ entry.item.badge > 99 ? '99+' : entry.item.badge }}
          </ProBadge>
        </NuxtLink>
      </template>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'

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
  | 'phone_in_talk'
  | 'record_voice_over'
  | 'event'
  | 'campaign'
  | 'slideshow'
  | 'analytics'
  | 'hub'
  | 'account_tree'
  | 'support_agent'
  | 'picture_as_pdf'
  | 'groups'
  | 'checklist'
  | 'receipt'
  | 'medication'
  | 'inventory_2'
  | 'local_shipping'
  | 'clinical_notes'

export type ProNavItem = {
  to: string
  label: string
  exact?: boolean
  icon: ProNavIcon
  badge?: number
  /** Static label badge (e.g. `dev`) — takes precedence over numeric badge. */
  tag?: string
  /** When set, starts a labeled group (shown once when consecutive items share the same label). */
  section?: string
}

const props = defineProps<{
  items: ProNavItem[]
}>()

type NavEntry = {
  key: string
  item: ProNavItem
  sectionLabel?: string
}

const entries = computed<NavEntry[]>(() => {
  let prevSection: string | undefined
  return props.items.map((item, i) => {
    const section = item.section
    const showSection = Boolean(section) && section !== prevSection
    if (section) prevSection = section
    else prevSection = undefined
    return {
      key: `${item.to}-${i}`,
      item,
      sectionLabel: showSection ? section : undefined,
    }
  })
})

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
  phone_in_talk: 'phone_in_talk',
  record_voice_over: 'record_voice_over',
  event: 'event',
  campaign: 'campaign',
  slideshow: 'slideshow',
  analytics: 'analytics',
  hub: 'hub',
  account_tree: 'account_tree',
  support_agent: 'support_agent',
  picture_as_pdf: 'picture_as_pdf',
  groups: 'groups',
  checklist: 'checklist',
  receipt: 'receipt_long',
  medication: 'medication',
  inventory_2: 'inventory_2',
  local_shipping: 'local_shipping',
  clinical_notes: 'clinical_notes',
}

function iconName(name: ProNavIcon) {
  return icons[name] ?? icons.dashboard
}

function navTestId(to: string) {
  // Stable aliases for existing e2e selectors.
  if (to === '/calendar') return 'nav-calendar'
  if (to === '/clients') return 'nav-clients'
  if (to === '/pets') return 'nav-pets'
  if (to === '/messages') return 'nav-messages'
  if (to === '/requests') return 'nav-requests'
  if (to === '/invoicing') return 'nav-invoicing'
  if (to === '/ordonnances') return 'nav-ordonnances'
  if (to === '/medicaments') return 'nav-medicaments'
  if (to === '/stock') return 'nav-stock'
  if (to === '/daf') return 'nav-daf'
  if (to === '/commissions') return 'nav-commissions'
  const slug = to
    .replace(/^\//, '')
    .replace(/\//g, '-')
    .replace(/[^a-z0-9-]/gi, '')
    .toLowerCase()
  return slug ? `nav-${slug}` : undefined
}

function badgeTestId(to: string) {
  if (to === '/calendar') return 'nav-calendar-badge'
  if (to === '/clients') return 'nav-clients-badge'
  if (to === '/pets') return 'nav-pets-badge'
  if (to === '/messages') return 'nav-messages-badge'
  if (to === '/requests') return 'nav-requests-badge'
  return undefined
}

function tagTestId(to: string) {
  if (to === '/invoicing') return 'nav-invoicing-tag'
  if (to === '/ordonnances') return 'nav-ordonnances-dev-tag'
  if (to === '/medicaments') return 'nav-medicaments-dev-tag'
  if (to === '/stock') return 'nav-stock-dev-tag'
  if (to === '/daf') return 'nav-daf-dev-tag'
  return undefined
}
</script>

<style scoped>
.pro-sidebar__section {
  margin: 0;
  padding: 0.35rem 0.85rem 0.2rem;
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.45);
  line-height: 1.2;
}

.pro-sidebar__section--spaced {
  margin-top: 0.65rem;
  padding-top: 0.55rem;
  border-top: 1px solid rgba(255, 255, 255, 0.12);
}

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

.pro-sidebar__tag {
  margin-left: auto;
  flex-shrink: 0;
  padding: 0.1rem 0.35rem;
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: lowercase;
  line-height: 1.2;
}
</style>
