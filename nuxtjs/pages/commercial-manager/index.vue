<template>
  <div data-testid="manager-dashboard-page">
    <ProPageHeader
      :title="$t('manager.dashboard.title')"
      :subtitle="$t('manager.dashboard.subtitle')"
    />

    <template v-if="overview">
      <h3 class="pro-mb-md">{{ $t('manager.dashboard.teamSection') }}</h3>
      <div class="pro-grid-kpi pro-mb-lg">
        <ProKpi :value="overview.teamProspectsTotal" :label="$t('manager.dashboard.prospects')" />
        <ProKpi :value="overview.teamProspectsContacted" :label="$t('manager.dashboard.contacted')" />
        <ProKpi :value="overview.teamAppointmentsUpcoming" :label="$t('manager.dashboard.appointmentsUpcoming')" />
        <ProKpi :value="overview.teamProspectsConverted" :label="$t('manager.dashboard.converted')" />
        <ProKpi :value="formatRate(overview.conversionRateBps)" :label="$t('manager.dashboard.conversionRate')" />
        <ProKpi :value="overview.teamStaleInPipeline" :label="$t('manager.dashboard.stale')" />
        <ProKpi :value="formatCurrency(overview.teamMonthEarnedCents)" :label="$t('manager.dashboard.teamMonthEarned')" />
        <ProKpi :value="overview.directoryTotal" :label="$t('manager.dashboard.directory')" />
      </div>

      <ProCard class="pro-mb-lg">
        <ProTable :empty="!overview.team?.length" :empty-title="$t('manager.dashboard.teamEmpty')">
          <thead>
            <tr>
              <th>{{ $t('manager.dashboard.colCommercial') }}</th>
              <th>{{ $t('manager.dashboard.colVets') }}</th>
              <th>{{ $t('manager.dashboard.colContacts') }}</th>
              <th>{{ $t('manager.dashboard.colAppointments') }}</th>
              <th>{{ $t('manager.dashboard.colConverted') }}</th>
              <th>{{ $t('manager.dashboard.colStale') }}</th>
              <th>{{ $t('manager.dashboard.colMonth') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in overview.team" :key="m.userId" :data-testid="`manager-team-row-${m.userId}`">
              <td>{{ m.fullName }}<br><span class="pro-hint">{{ m.email }}</span></td>
              <td>{{ m.assignedVets }}</td>
              <td>{{ m.contacts30d }}</td>
              <td>{{ m.appointmentsUpcoming }} / {{ m.appointmentsDone }}</td>
              <td>{{ m.prospectsConverted }}</td>
              <td>{{ m.staleInPipeline }}</td>
              <td>{{ formatCurrency(m.monthEarnedCents) }}</td>
            </tr>
          </tbody>
        </ProTable>
      </ProCard>

      <h3 class="pro-mb-md">{{ $t('manager.dashboard.selfSection') }}</h3>
      <div class="pro-grid-kpi">
        <ProKpi :value="overview.self?.assignedVets ?? 0" :label="$t('commercial.dashboard.assignedVets')" />
        <ProKpi :value="overview.self?.prospectsTotal ?? 0" :label="$t('commercial.dashboard.prospects')" />
        <ProKpi :value="overview.self?.prospectsConverted ?? 0" :label="$t('commercial.dashboard.converted')" />
        <ProKpi :value="formatCurrency(overview.self?.monthEarnedCents ?? 0)" :label="$t('commercial.dashboard.monthEarned')" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const { formatCurrency } = useFormatters()
const overview = ref<any>(null)

function formatRate(bps: number) {
  if (!bps) return '—'
  return `${(bps / 100).toFixed(1)} %`
}

onMounted(async () => {
  const res: any = await $fetch('/api/commercial-manager/overview')
  overview.value = res.data ?? res
})
</script>
