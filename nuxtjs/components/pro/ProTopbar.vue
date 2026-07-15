<template>
  <header class="pro-topbar" data-testid="pro-topbar">
    <div class="pro-topbar__left">
      <PetsFollowLogo variant="compact" :link-to="homeLink" />
      <slot name="breadcrumb" />
    </div>
    <div class="pro-topbar__actions">
      <div v-if="showNotifications" class="pro-topbar__dropdown-wrap">
        <button
          type="button"
          class="pro-topbar__icon-btn"
          :aria-label="$t('components.topbar.notifications')"
          aria-haspopup="true"
          :aria-expanded="notifOpen"
          data-testid="pro-notifications-btn"
          @click="toggleNotif"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path
              d="M12 3a5 5 0 00-5 5v2.5c0 .9-.3 1.8-.9 2.5L5 15.5h14l-1.1-2.5c-.6-.7-.9-1.6-.9-2.5V8a5 5 0 00-5-5z"
              stroke="currentColor"
              stroke-width="1.75"
              stroke-linejoin="round"
            />
            <path d="M10 18a2 2 0 004 0" stroke="currentColor" stroke-width="1.75" stroke-linecap="round"/>
          </svg>
          <span v-if="notifCount > 0" class="pro-topbar__badge">{{ notifCount }}</span>
        </button>
        <div v-if="notifOpen" class="pro-topbar__dropdown" role="menu">
          <p class="pro-topbar__dropdown-title">{{ $t('components.topbar.notifications') }}</p>
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

      <div class="pro-topbar__dropdown-wrap">
        <button
          type="button"
          class="pro-topbar__profile-btn"
          :aria-label="$t('components.topbar.profileMenu')"
          aria-haspopup="true"
          :aria-expanded="profileOpen"
          data-testid="pro-profile-btn"
          @click="toggleProfile"
        >
          <span class="pro-avatar">{{ userInitials }}</span>
          <span class="pro-topbar__profile-name">{{ userName }}</span>
        </button>
        <div v-if="profileOpen" class="pro-topbar__dropdown pro-topbar__dropdown--profile" role="menu">
          <p class="pro-topbar__dropdown-title">{{ userName }}</p>
          <p class="pro-topbar__dropdown-email">{{ userEmail }}</p>
          <NuxtLink
            v-if="settingsLink"
            :to="settingsLink"
            class="pro-topbar__dropdown-link"
            @click="profileOpen = false"
          >
            {{ $t('components.topbar.settings') }}
          </NuxtLink>
          <button type="button" class="pro-topbar__logout-btn" @click="handleLogout">
            {{ $t('common.logout') }}
          </button>
        </div>
      </div>
    </div>
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
const { user, fetchUser, initials, logout } = useProUser()
const { items: notifItems, count: notifCount, refresh: refreshNotif } = useProNotifications()

const notifOpen = ref(false)
const profileOpen = ref(false)

const userName = computed(() => user.value?.fullName || t('common.user'))
const userEmail = computed(() => user.value?.email || '')
const userInitials = computed(() => initials())

onMounted(async () => {
  document.addEventListener('click', onDocClick)
  await fetchUser()
  if (props.showNotifications) await refreshNotif()
})

onUnmounted(() => document.removeEventListener('click', onDocClick))

function toggleNotif() {
  notifOpen.value = !notifOpen.value
  profileOpen.value = false
}

function toggleProfile() {
  profileOpen.value = !profileOpen.value
  notifOpen.value = false
}

function handleLogout() {
  profileOpen.value = false
  logout()
}

function onDocClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.pro-topbar__dropdown-wrap')) {
    notifOpen.value = false
    profileOpen.value = false
  }
}
</script>
