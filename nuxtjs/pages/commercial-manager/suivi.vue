<template>
  <div data-testid="manager-followups-page">
    <ProPageHeader
      :title="$t('manager.followups.title')"
      :subtitle="$t('manager.followups.subtitle')"
    />

    <ProCard class="pro-mb-lg">
      <h3 class="pro-mb-md">{{ $t('manager.followups.upcomingTitle') }}</h3>
      <ProTable :empty="!upcoming.length" :empty-title="$t('manager.followups.upcomingEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('manager.followups.colCommercial') }}</th>
            <th>{{ $t('commercial.prospects.appointmentAt') }}</th>
            <th>{{ $t('commercial.prospects.statusLabel') }}</th>
            <th>{{ $t('commercial.prospects.city') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in upcoming" :key="p.id">
            <td>{{ p.practiceName }}</td>
            <td>{{ p.commercialName || p.commercialEmail || '—' }}</td>
            <td>{{ formatDt(p.appointmentAt) }}</td>
            <td>{{ $t(`commercial.prospects.status.${p.status}`) }}</td>
            <td>{{ p.city }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>

    <ProCard>
      <h3 class="pro-mb-md">{{ $t('manager.followups.staleTitle') }}</h3>
      <ProTable :empty="!stale.length" :empty-title="$t('manager.followups.staleEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.prospects.practiceName') }}</th>
            <th>{{ $t('manager.followups.colCommercial') }}</th>
            <th>{{ $t('commercial.prospects.statusLabel') }}</th>
            <th>{{ $t('commercial.prospects.daysInStatus') }}</th>
            <th>{{ $t('commercial.prospects.city') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in stale" :key="p.id">
            <td>{{ p.practiceName }}</td>
            <td>{{ p.commercialName || p.commercialEmail || '—' }}</td>
            <td>{{ $t(`commercial.prospects.status.${p.status}`) }}</td>
            <td>{{ p.daysInStatus }}</td>
            <td>{{ p.city }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const upcoming = ref<any[]>([])
const stale = ref<any[]>([])

function formatDt(v?: string) {
  if (!v) return '—'
  try {
    return new Date(v).toLocaleString()
  } catch {
    return v
  }
}

onMounted(async () => {
  const res: any = await $fetch('/api/commercial-manager/followups')
  const data = res.data ?? res ?? {}
  upcoming.value = data.upcomingAppointments ?? []
  stale.value = data.staleProspects ?? []
})
</script>
