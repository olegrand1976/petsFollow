<template>
  <header class="pro-topbar" data-testid="pro-topbar">
    <div class="pro-topbar__left">
      <PetsFollowLogo variant="compact" :link-to="homeLink" />
      <span
        v-if="practiceName"
        class="pro-topbar__practice"
        data-testid="pro-topbar-practice"
      >{{ practiceName }}</span>
      <slot name="breadcrumb" />
    </div>
    <div class="pro-topbar__actions">
      <button
        type="button"
        class="pro-topbar__icon-btn"
        :aria-label="$t('support.buttonAria')"
        data-testid="pro-support-btn"
        @click.stop="openSupport"
      >
        <ProIcon name="support_agent" :size="20" />
      </button>
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
          <NuxtLink
            v-if="settingsLink"
            :to="settingsLink"
            class="pro-topbar__dropdown-link"
            @click="closeProfileDetails"
          >
            {{ $t('components.topbar.settings') }}
          </NuxtLink>
          <button
            type="button"
            class="pro-topbar__logout-btn"
            data-testid="pro-logout-btn"
            @click="handleLogout"
          >
            {{ $t('common.logout') }}
          </button>
        </div>
      </details>
    </div>
    <ProSupportDialog v-model:open="supportOpen" />
  </header>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    homeLink?: string
    settingsLink?: string
    showNotifications?: boolean
  }>(),
  {
    homeLink: '/',
    settingsLink: undefined,
    showNotifications: true,
  },
)

const { t } = useI18n()
const { isDark, toggleTheme } = useColorTheme()
const { user, fetchUser, logout } = useProUser()
const {
  items: notifItems,
  count: notifCount,
  refresh: refreshNotif,
  markAllRead,
  startPolling,
  stopPolling,
} = useProNotifications()

const notifOpen = ref(false)
const supportOpen = ref(false)

const userName = computed(() => user.value?.fullName || t('common.user'))
const userEmail = computed(() => user.value?.email || '')
const practiceName = computed(() => user.value?.practiceName?.trim() || '')

onMounted(async () => {
  document.addEventListener('click', onDocClick)
  try {
    await fetchUser()
  } catch { /* 401 handled by middleware */ }
  if (props.showNotifications) {
    await refreshNotif()
    startPolling()
  }
})

onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
  if (props.showNotifications) stopPolling()
})

function toggleNotif() {
  notifOpen.value = !notifOpen.value
  closeProfileDetails()
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

function openSupport() {
  closeProfileDetails()
  notifOpen.value = false
  // Defer open so the triggering click cannot hit the modal backdrop and close it.
  nextTick(() => {
    supportOpen.value = true
  })
}

function handleLogout() {
  closeProfileDetails()
  logout()
}

function onDocClick(e: MouseEvent) {
  const target = e.target
  if (!(target instanceof Element)) return
  if (!target.closest('.pro-topbar__dropdown-wrap')) {
    notifOpen.value = false
    closeProfileDetails()
  }
}
</script>
