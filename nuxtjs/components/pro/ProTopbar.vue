<template>
  <header class="pro-topbar" data-testid="pro-topbar">
    <div class="pro-topbar__left">
      <button
        type="button"
        class="pro-topbar__icon-btn pro-topbar__menu-btn"
        :aria-label="navOpen ? $t('components.topbar.closeNav') : $t('components.topbar.openNav')"
        :aria-expanded="navOpen"
        aria-controls="pro-nav-drawer"
        data-testid="pro-nav-menu-btn"
        @click="onToggleNav"
      >
        <ProIcon :name="navOpen ? 'close' : 'menu'" :size="20" />
      </button>
      <PetsFollowLogo variant="compact" :link-to="homeLink" />
      <span
        v-if="showStagingTag"
        class="pro-topbar__env"
        data-testid="pro-topbar-staging"
        :title="$t('common.stagingEnv')"
        :aria-label="$t('common.stagingEnv')"
      >S</span>
      <span
        v-if="practiceName"
        class="pro-topbar__practice"
        data-testid="pro-topbar-practice"
      >{{ practiceName }}</span>
      <label
        v-if="showSiteSwitcher"
        class="pro-topbar__site"
        data-testid="pro-site-switcher"
      >
        <select
          class="pro-topbar__site-select"
          :value="effectiveSiteId"
          :aria-label="$t('sites.switcherLabel')"
          data-testid="pro-site-select"
          @change="onSiteChange"
        >
          <option
            v-for="s in sites"
            :key="s.id"
            :value="s.id"
          >{{ s.name }}</option>
          <option :value="SITE_ALL">{{ $t('sites.allSites') }}</option>
        </select>
      </label>
      <slot name="breadcrumb" />
    </div>
    <div class="pro-topbar__actions">
      <NuxtLink
        to="/support"
        class="pro-topbar__icon-btn"
        :aria-label="$t('support.buttonAria')"
        data-testid="pro-support-btn"
        @click.capture="onSupportNav"
      >
        <ProIcon name="support_agent" :size="20" />
      </NuxtLink>
      <ProDeskSwitcher v-if="showDeskSwitcher" />
      <ProLocaleSelect persist />
      <button
        type="button"
        class="pro-topbar__icon-btn"
        :aria-label="isDark ? $t('components.topbar.themeLight') : $t('components.topbar.themeDark')"
        data-testid="pro-theme-toggle"
        @click="toggleTheme"
      >
        <ProIcon :name="isDark ? 'light_mode' : 'dark_mode'" :size="20" />
      </button>

      <div v-if="showNotifications" class="pro-topbar__dropdown-wrap">
        <button
          type="button"
          class="pro-topbar__icon-btn"
          :aria-label="$t('components.topbar.notifications')"
          aria-haspopup="true"
          :aria-expanded="notifOpen"
          data-testid="pro-notifications-btn"
          @click.stop="toggleNotif"
        >
          <ProIcon name="notifications" :size="20" />
          <span v-if="notifCount > 0" class="pro-topbar__badge">{{ notifCount }}</span>
        </button>
        <div v-if="notifOpen" class="pro-topbar__dropdown" role="menu">
          <div class="pro-topbar__dropdown-header">
            <p class="pro-topbar__dropdown-title">{{ $t('components.topbar.notifications') }}</p>
            <button
              v-if="notifCount > 0"
              type="button"
              class="pro-topbar__mark-all"
              data-testid="pro-notifications-mark-all"
              @click="handleMarkAllRead"
            >
              {{ $t('components.topbar.markAllRead') }}
            </button>
          </div>
          <ProEmptyState
            v-if="!notifItems.length"
            :title="$t('components.topbar.notificationsEmptyTitle')"
            :description="$t('components.topbar.notificationsEmptyDescription')"
          />
          <ul v-else class="pro-topbar__notif-list">
            <li v-for="item in notifItems" :key="item.id">
              <NuxtLink :to="item.href" @click="notifOpen = false">
                <strong>{{ item.label }}</strong>
                <span v-if="item.preview" class="pro-topbar__notif-preview">{{ item.preview }}</span>
              </NuxtLink>
            </li>
          </ul>
          <NuxtLink to="/messages" class="pro-topbar__dropdown-link" @click="notifOpen = false">
            {{ $t('common.seeAll') }}
          </NuxtLink>
        </div>
      </div>

      <!-- details/summary: opens without Vue hydration (Cloud Run SSR / e2e). -->
      <details class="pro-topbar__dropdown-wrap" data-testid="pro-profile-details">
        <summary
          class="pro-topbar__profile-btn"
          :aria-label="$t('components.topbar.profileMenu')"
          data-testid="pro-profile-btn"
          @click="notifOpen = false"
        >
          <ProAvatar :src="user?.avatarUrl" :name="userName" size="sm" />
          <span class="pro-topbar__profile-name">{{ userName }}</span>
        </summary>
        <div class="pro-topbar__dropdown pro-topbar__dropdown--profile" role="menu">
          <p class="pro-topbar__dropdown-title">{{ userName }}</p>
          <p class="pro-topbar__dropdown-email">{{ userEmail }}</p>
          <div
            v-if="switchableProfiles.length > 1"
            class="pro-topbar__profiles"
            data-testid="pro-profile-switcher"
          >
            <p class="pro-topbar__dropdown-section">{{ $t('components.topbar.switchProfile') }}</p>
            <button
              v-for="p in switchableProfiles"
              :key="p.id"
              type="button"
              class="pro-topbar__profile-switch"
              :class="{ 'is-active': p.active }"
              :disabled="p.active || profileSwitchBusy"
              :data-testid="`pro-profile-switch-${p.role}`"
              @click="switchToProfile(p)"
            >
              <span>{{ profileRoleLabel(p.role) }}</span>
              <ProBadge v-if="p.active" variant="success">{{ $t('components.topbar.activeProfile') }}</ProBadge>
            </button>
            <p v-if="profileSwitchError" class="pro-topbar__profile-error" role="alert">{{ profileSwitchError }}</p>
          </div>
          <NuxtLink
            v-if="settingsLink"
            :to="settingsLink"
            class="pro-topbar__dropdown-link"
            @click="closeProfileDetails"
          >
            {{ $t('components.topbar.settings') }}
          </NuxtLink>
          <a
            href="/api/auth/logout-redirect"
            class="pro-topbar__logout-btn"
            data-testid="pro-logout-btn"
            @click="closeProfileDetails"
          >
            {{ $t('common.logout') }}
          </a>
        </div>
      </details>
    </div>
  </header>
</template>

<script setup lang="ts">
import { homePathForRole, isPracticeStaffRole, isProRole } from '~/composables/useAuth'
import { canActivateProfile, homeSwitchRole, profileRoleSortIndex } from '~/utils/profile-switch'

type ProfileRow = {
  id: string
  role: string
  active?: boolean
  practiceId?: string
  createdAt?: string
}

const props = withDefaults(
  defineProps<{
    homeLink?: string
    settingsLink?: string
    showNotifications?: boolean
    showDeskSwitcher?: boolean
  }>(),
  {
    homeLink: '/',
    settingsLink: undefined,
    showNotifications: true,
    showDeskSwitcher: true,
  },
)

const { t } = useI18n()
const { isDark, toggleTheme } = useColorTheme()
const { user, fetchUser } = useProUser()
const {
  SITE_ALL,
  sites,
  multiSite,
  sitesUiEnabled,
  effectiveSiteId,
  initFromStorage,
  setSiteId,
} = usePracticeSites()
const showSiteSwitcher = computed(
  () => sitesUiEnabled.value && isPracticeStaffRole(user.value?.role) && multiSite.value,
)
function onSiteChange(ev: Event) {
  const v = (ev.target as HTMLSelectElement)?.value || ''
  setSiteId(v)
  if (import.meta.client) {
    window.dispatchEvent(new CustomEvent('pf-site-changed', { detail: { siteId: v } }))
  }
}
const { isStaging } = useAppEnv()
const { captureOriginPage } = useSupportDiagnostics()
const desk = useDeskSession()
const { open: navOpen, toggleDrawer, closeDrawer } = useProNavDrawer()
const {
  items: notifItems,
  count: notifCount,
  refresh: refreshNotif,
  markAllRead,
  startPolling,
  stopPolling,
} = useProNotifications()

const notifOpen = ref(false)
const profiles = ref<ProfileRow[]>([])
const profileSwitchBusy = ref(false)
const profileSwitchError = ref('')

function onToggleNav() {
  notifOpen.value = false
  closeProfileDetails()
  toggleDrawer()
}

const userName = computed(() => user.value?.fullName || t('common.user'))
const userEmail = computed(() => user.value?.email || '')
const practiceName = computed(() => user.value?.practiceName?.trim() || '')
/** Tag « S » : environnement staging + session authentifiée. */
const showStagingTag = computed(() => isStaging.value && !!user.value)
const showDeskSwitcher = computed(
  () => props.showDeskSwitcher !== false && isPracticeStaffRole(user.value?.role),
)
/** Nuxt Pro : profils face Pro, filtrés par matrice CanActivate (pas client Flutter). */
const switchableProfiles = computed(() => {
  const home = homeSwitchRole(profiles.value)
  return profiles.value
    .filter((p) => isProRole(p.role) && canActivateProfile(home, p.role))
    .slice()
    .sort((a, b) => {
      if (a.active && !b.active) return -1
      if (!a.active && b.active) return 1
      return profileRoleSortIndex(a.role) - profileRoleSortIndex(b.role)
    })
})

onMounted(async () => {
  initFromStorage()
  document.addEventListener('click', onDocClick)
  try {
    await fetchUser()
    initFromStorage()
  } catch { /* 401 handled by middleware */ }
  void loadProfiles()
  // Idle/bootstrap owned by layout — topbar only renders switcher.
  if (props.showNotifications && !desk.uiBlocked.value) {
    await refreshNotif()
    startPolling()
  }
})

onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
  if (props.showNotifications) stopPolling()
})

watch(
  () => desk.uiBlocked.value,
  async (blocked) => {
    if (!props.showNotifications) return
    if (blocked) {
      stopPolling()
      return
    }
    await refreshNotif()
    startPolling()
  },
)

function toggleNotif() {
  notifOpen.value = !notifOpen.value
  closeProfileDetails()
  closeDrawer()
  if (notifOpen.value) void refreshNotif()
}

async function handleMarkAllRead() {
  try {
    await markAllRead()
  } catch {
    // keep dropdown open; badge will refresh on next open
  }
}

function closeProfileDetails() {
  const el = document.querySelector('[data-testid="pro-profile-details"]') as HTMLDetailsElement | null
  if (el) el.open = false
}

function onSupportNav() {
  captureOriginPage()
  closeProfileDetails()
  notifOpen.value = false
}

function onDocClick(e: MouseEvent) {
  const target = e.target
  if (!(target instanceof Element)) return
  if (!target.closest('.pro-topbar__dropdown-wrap')) {
    notifOpen.value = false
    closeProfileDetails()
  }
}

async function loadProfiles() {
  try {
    const res: any = await $fetch('/api/me/profiles')
    const data = res?.data ?? res
    profiles.value = Array.isArray(data) ? data : (data?.profiles ?? data?.items ?? [])
  } catch {
    profiles.value = []
  }
}

function profileRoleLabel(role: string) {
  const key = `components.topbar.profileRole_${role}`
  const label = t(key)
  return label === key ? role : label
}

async function switchToProfile(p: ProfileRow) {
  if (p.active || profileSwitchBusy.value) return
  profileSwitchBusy.value = true
  profileSwitchError.value = ''
  try {
    await $fetch('/api/me/profiles/switch', {
      method: 'POST',
      body: { profileId: p.id },
    })
    closeProfileDetails()
    await fetchUser().catch(() => null)
    await loadProfiles()
    const role = user.value?.role || p.role
    await navigateTo(homePathForRole(role, { profileComplete: user.value?.profileComplete }))
  } catch {
    profileSwitchError.value = t('components.topbar.switchProfileError')
  } finally {
    profileSwitchBusy.value = false
  }
}
</script>

