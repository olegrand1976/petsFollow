<template>
  <div>
    <ProPageHeader
      :title="welcomeTitle"
      :subtitle="$t('dashboard.subtitle')"
    />
    <div class="pro-grid-kpi">
      <ProCard>
        <div class="pro-kpi pro-kpi--with-icon">
          <ProIcon name="group" class="pro-kpi__icon" :size="20" />
          <span class="pro-kpi__value">{{ clientCount }}</span>
          <span class="pro-kpi__label">{{ $t('dashboard.activeClients') }}</span>
        </div>
      </ProCard>
      <ProCard>
        <div class="pro-kpi pro-kpi--with-icon">
          <ProIcon name="chat" class="pro-kpi__icon" :size="20" />
          <span class="pro-kpi__value">{{ unreadCount }}</span>
          <span class="pro-kpi__label">{{ $t('dashboard.unreadMessages') }}</span>
        </div>
      </ProCard>
      <ProCard>
        <div class="pro-kpi pro-kpi--with-icon">
          <ProIcon name="favorite" class="pro-kpi__icon" :size="20" />
          <span class="pro-kpi__value">{{ recentSessions }}</span>
          <span class="pro-kpi__label">{{ $t('dashboard.recentSessions') }}</span>
        </div>
      </ProCard>
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
const { fetchUser } = useProUser()

onMounted(async () => {
  const me = await fetchUser()
  const name = me?.fullName
  if (name) welcomeTitle.value = t('dashboard.welcome', { name: name.split(' ')[0] })
  try {
    const res: any = await $fetch('/api/vet/overview')
    const data = res.data ?? res
    clientCount.value = String(data.clientCount ?? 0)
    unreadCount.value = String(data.unreadMessages ?? 0)
    recentSessions.value = String(data.recentSessions7d ?? 0)
  } catch {
    clientCount.value = '0'
    unreadCount.value = '0'
    recentSessions.value = '0'
  }
})
</script>
