<template>
  <div data-testid="commercial-dashboard-page">
    <ProPageHeader
      :title="$t('commercial.dashboard.title')"
      :subtitle="$t('commercial.dashboard.subtitle')"
    >
      <template #actions>
        <ProButton
          variant="secondary"
          test-id="commercial-app-invite-open"
          @click="appInviteOpen = true"
        >
          <ProIcon name="qr_code_2" />
          {{ $t('commercial.appInvite.open') }}
        </ProButton>
      </template>
    </ProPageHeader>
    <ProAppInviteModal v-model:open="appInviteOpen" :title="$t('commercial.appInvite.title')" />

    <ProEmptyState v-if="loadError" :title="$t('commercial.dashboard.loadError')" />
    <p v-else-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>

    <template v-else-if="overview">
      <div class="pro-grid-kpi pro-mb-lg">
        <ProKpi :value="overview.assignedVets" :label="$t('commercial.dashboard.assignedVets')" />
        <ProKpi :value="overview.referredClients ?? 0" :label="$t('commercial.dashboard.referredClients')" />
        <ProKpi :value="overview.prospectsTotal" :label="$t('commercial.dashboard.prospects')" />
        <ProKpi :value="overview.directoryProspects ?? 0" :label="$t('commercial.dashboard.directory')" />
        <ProKpi :value="overview.prospectsNew" :label="$t('commercial.dashboard.prospectsNew')" />
        <ProKpi :value="overview.prospectsConverted" :label="$t('commercial.dashboard.converted')" />
        <ProKpi :value="overview.paidPetsMonth ?? 0" :label="$t('commercial.dashboard.paidPets')" />
        <ProKpi :value="formatMix(overview.triennialMixBps)" :label="$t('commercial.dashboard.mix')" />
        <ProKpi :value="overview.appointmentsUpcoming ?? 0" :label="$t('commercial.dashboard.upcoming')" />
        <ProKpi :value="overview.staleInPipeline ?? 0" :label="$t('commercial.dashboard.stale')" />
        <ProKpi :value="formatCurrency(overview.monthEarnedCents)" :label="$t('commercial.dashboard.monthEarned')" />
        <ProKpi :value="formatCurrency(overview.lifetimeEarnedCents)" :label="$t('commercial.dashboard.lifetimeEarned')" />
        <ProKpi :value="formatCurrency(overview.linkedSubscriptionRevenueCents)" :label="$t('commercial.dashboard.subRevenue')" />
        <ProKpi :value="formatCurrency(overview.linkedAddonRevenueCents)" :label="$t('commercial.dashboard.addonRevenue')" />
      </div>

      <div class="pf-dash-cards">
        <ProCard class="pro-mb-lg">
          <div class="pf-card-head">
            <h3>{{ $t('commercial.dashboard.upcomingTitle') }}</h3>
            <NuxtLink to="/commercial/prospects" class="pro-link">{{ $t('commercial.prospects.title') }}</NuxtLink>
          </div>
          <ProTable
            :empty="!upcoming.length"
            :empty-title="$t('commercial.dashboard.upcomingEmpty')"
          >
            <thead>
              <tr>
                <th>{{ $t('commercial.prospects.practiceName') }}</th>
                <th>{{ $t('commercial.prospects.appointmentAt') }}</th>
                <th>{{ $t('commercial.prospects.city') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in upcoming" :key="p.id">
                <td>{{ p.practiceName }}</td>
                <td>{{ formatDt(p.appointmentAt) }}</td>
                <td>{{ p.city || '—' }}</td>
              </tr>
            </tbody>
          </ProTable>
        </ProCard>

        <ProCard class="pro-mb-lg">
          <div class="pf-card-head">
            <h3>{{ $t('commercial.dashboard.staleTitle') }}</h3>
            <NuxtLink to="/commercial/prospects" class="pro-link">{{ $t('commercial.prospects.title') }}</NuxtLink>
          </div>
          <ProTable
            :empty="!stale.length"
            :empty-title="$t('commercial.dashboard.staleEmpty')"
          >
            <thead>
              <tr>
                <th>{{ $t('commercial.prospects.practiceName') }}</th>
                <th>{{ $t('commercial.prospects.statusLabel') }}</th>
                <th>{{ $t('commercial.prospects.daysInStatus') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in stale" :key="p.id">
                <td>{{ p.practiceName }}</td>
                <td>{{ $t(`commercial.prospects.status.${p.status}`) }}</td>
                <td>{{ p.daysInStatus }}</td>
              </tr>
            </tbody>
          </ProTable>
        </ProCard>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { formatCurrency, formatDate } = useFormatters()
const overview = ref<any>(null)
const loading = ref(true)
const loadError = ref(false)
const appInviteOpen = ref(false)

const upcoming = computed(() => overview.value?.upcomingAppointments ?? [])
const stale = computed(() => overview.value?.staleProspects ?? [])

function formatMix(bps?: number) {
  if (bps == null || bps === 0) return '—'
  return `${(bps / 100).toFixed(1)} %`
}

function formatDt(v?: string) {
  if (!v) return '—'
  try {
    return formatDate(v)
  } catch {
    return v
  }
}

onMounted(async () => {
  loading.value = true
  loadError.value = false
  try {
    const res: any = await $fetch('/api/commercial/overview')
    overview.value = res.data ?? res
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.pf-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}
.pf-card-head h3 {
  margin: 0;
}
.pro-link {
  color: var(--pf-vet-accent);
  text-decoration: none;
  font-size: 0.875rem;
}
.pro-link:hover {
  text-decoration: underline;
}
</style>
