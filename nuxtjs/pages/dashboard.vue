<template>
  <div data-testid="vet-dashboard-page">
    <ProPageHeader
      :title="welcomeTitle"
      :subtitle="$t('dashboard.subtitle')"
    />
    <div class="pro-grid-kpi">
      <ProKpi
        icon="group"
        :value="clientCount"
        :label="$t('dashboard.activeClients')"
        to="/clients"
      />
      <ProKpi
        icon="chat"
        :value="unreadCount"
        :label="$t('dashboard.unreadMessages')"
        to="/messages"
        :variant="hasUnread ? 'alert' : 'default'"
      />
      <ProKpi
        icon="favorite"
        :value="recentSessions"
        :label="$t('dashboard.recentSessions')"
      />
    </div>
    <div class="pro-grid-2 pro-mt-lg">
      <ProCard :title="$t('dashboard.quickActions')">
        <div class="pro-flex-gap">
          <ProButton @click="navigateTo('/clients')">{{ $t('dashboard.viewClients') }}</ProButton>
          <ProButton variant="secondary" @click="navigateTo('/messages')">{{ $t('dashboard.messaging') }}</ProButton>
          <ProButton variant="ghost" @click="navigateTo('/settings')">{{ $t('nav.settings') }}</ProButton>
        </div>
      </ProCard>
      <ProCard :title="$t('dashboard.about')">
        <p class="text-muted">{{ $t('dashboard.aboutText') }}</p>
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const { t } = useI18n()
const welcomeTitle = ref(t('dashboard.title'))
const clientCount = ref('—')
const unreadCount = ref('—')
const recentSessions = ref('—')
const unreadRaw = ref(0)
const { fetchUser } = useProUser()

const hasUnread = computed(() => unreadRaw.value > 0)

onMounted(async () => {
  const me = await fetchUser()
  const name = me?.fullName
  if (name) welcomeTitle.value = t('dashboard.welcome', { name: name.split(' ')[0] })
  try {
    const res: any = await $fetch('/api/vet/overview')
    const data = res.data ?? res
    clientCount.value = String(data.clientCount ?? 0)
    unreadRaw.value = Number(data.unreadMessages ?? 0)
    unreadCount.value = String(unreadRaw.value)
    recentSessions.value = String(data.recentSessions7d ?? 0)
  } catch {
    clientCount.value = '0'
    unreadCount.value = '0'
    unreadRaw.value = 0
    recentSessions.value = '0'
  }
})
</script>
