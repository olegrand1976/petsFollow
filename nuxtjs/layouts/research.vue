<template>
  <div class="pro-app">
    <template v-if="shellReady">
      <ProTopbar home-link="/research" settings-link="/research/settings" :show-notifications="false" />
      <div class="pro-app-shell">
        <ProSidebar :items="navItems" />
        <div class="pro-app-body">
          <main class="pro-main main">
            <div class="pro-main-inner">
              <slot />
            </div>
          </main>
        </div>
      </div>
    </template>
    <main v-else class="pro-main main">
      <div class="pro-main-inner">
        <slot />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import type { ProNavItem } from '~/components/pro/ProSidebar.vue'
import { isPublicFlagOn } from '~/utils/public-feature-flag'

const { t } = useI18n()
const { user, fetchUser } = useProUser()
const runtimeConfig = useRuntimeConfig()
if (!user.value) {
  await fetchUser().catch(() => null)
}
const shellReady = computed(() => !!user.value?.role)
const researchOn = computed(() => isPublicFlagOn(runtimeConfig.public.researchEnabled))

const navItems = computed<ProNavItem[]>(() => {
  const tagDev = t('nav.tagDev')
  const items: ProNavItem[] = [
    {
      to: '/research',
      label: t('nav.researchOverview'),
      exact: true,
      icon: 'analytics',
      section: t('nav.section.research'),
      tag: tagDev,
    },
    {
      to: '/research/heatmap',
      label: t('nav.researchHeatmap'),
      icon: 'hub',
      section: t('nav.section.research'),
      tag: tagDev,
    },
    {
      to: '/research/timeseries',
      label: t('nav.researchTimeseries'),
      icon: 'event',
      section: t('nav.section.research'),
      tag: tagDev,
    },
    {
      to: '/research/alerts',
      label: t('nav.researchAlerts'),
      icon: 'campaign',
      section: t('nav.section.research'),
      tag: tagDev,
    },
    {
      to: '/research/groups',
      label: t('nav.researchGroups'),
      icon: 'groups',
      section: t('nav.section.research'),
      tag: tagDev,
    },
    {
      to: '/research/dataroom',
      label: t('nav.researchDataroom'),
      icon: 'description',
      section: t('nav.section.research'),
      tag: tagDev,
    },
  ]
  return researchOn.value ? items : items.slice(0, 1)
})

onMounted(() => {
  if (!user.value) void fetchUser().catch(() => {})
})
</script>
