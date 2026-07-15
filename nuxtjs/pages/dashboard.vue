<template>
  <div>
    <ProPageHeader
      :title="welcomeTitle"
      subtitle="Vue d'ensemble de votre activité petsFollow Pro."
    />
    <div class="pro-grid-kpi">
      <ProCard>
        <div class="pro-kpi pro-kpi--with-icon">
          <span class="pro-kpi__icon" aria-hidden="true">👥</span>
          <span class="pro-kpi__value">{{ clientCount }}</span>
          <span class="pro-kpi__label">Clients actifs</span>
        </div>
      </ProCard>
      <ProCard>
        <div class="pro-kpi pro-kpi--with-icon">
          <span class="pro-kpi__icon" aria-hidden="true">💬</span>
          <span class="pro-kpi__value">{{ unreadCount }}</span>
          <span class="pro-kpi__label">Messages non lus</span>
        </div>
      </ProCard>
      <ProCard>
        <div class="pro-kpi pro-kpi--with-icon">
          <span class="pro-kpi__icon" aria-hidden="true">❤️</span>
          <span class="pro-kpi__value">{{ recentSessions }}</span>
          <span class="pro-kpi__label">Relevés récents (7 j)</span>
        </div>
      </ProCard>
    </div>
    <div class="pro-grid-2 pro-mt-lg">
      <ProCard title="Actions rapides">
        <div class="pro-flex-gap">
          <ProButton @click="navigateTo('/clients')">Voir les clients</ProButton>
          <ProButton variant="secondary" @click="navigateTo('/messages')">Messagerie</ProButton>
          <ProButton variant="ghost" @click="navigateTo('/settings')">Paramètres</ProButton>
        </div>
      </ProCard>
      <ProCard title="À propos">
        <p class="text-muted">
          petsFollow Pro centralise le suivi cardiaque, les dossiers clients et la messagerie
          sécurisée pour votre cabinet vétérinaire.
        </p>
      </ProCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'vet-only' })

const welcomeTitle = ref('Dashboard')
const clientCount = ref('—')
const unreadCount = ref('—')
const recentSessions = ref('—')
const { fetchUser } = useProUser()

onMounted(async () => {
  const me = await fetchUser()
  const name = me?.fullName
  if (name) welcomeTitle.value = `Bonjour, ${name.split(' ')[0]}`
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
