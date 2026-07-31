<template>
  <div data-testid="manager-member-page">
    <ProPageHeader
      :title="title"
      :subtitle="$t('manager.member.subtitle')"
    >
      <template #actions>
        <NuxtLink to="/commercial-manager" class="pro-link">
          {{ $t('manager.member.back') }}
        </NuxtLink>
      </template>
    </ProPageHeader>

    <ProEmptyState v-if="loadError" :title="$t('manager.member.loadError')" />
    <p v-else-if="loading" class="pro-hint">{{ $t('common.loading') }}</p>

    <template v-else-if="overview">
      <div class="pro-grid-kpi">
        <ProKpi :value="overview.assignedVets ?? 0" :label="$t('commercial.dashboard.assignedVets')" />
        <ProKpi :value="overview.prospectsTotal ?? 0" :label="$t('commercial.dashboard.prospects')" />
        <ProKpi :value="overview.prospectsNew ?? 0" :label="$t('commercial.dashboard.prospectsNew')" />
        <ProKpi :value="overview.prospectsConverted ?? 0" :label="$t('commercial.dashboard.converted')" />
        <ProKpi :value="overview.paidPetsMonth ?? 0" :label="$t('commercial.dashboard.paidPets')" />
        <ProKpi :value="formatMix(overview.triennialMixBps)" :label="$t('commercial.dashboard.mix')" />
        <ProKpi :value="overview.appointmentsUpcoming ?? 0" :label="$t('commercial.dashboard.upcoming')" />
        <ProKpi :value="overview.staleInPipeline ?? 0" :label="$t('commercial.dashboard.stale')" />
        <ProKpi :value="formatCurrency(overview.monthEarnedCents ?? 0)" :label="$t('commercial.dashboard.monthEarned')" />
        <ProKpi :value="formatCurrency(overview.lifetimeEarnedCents ?? 0)" :label="$t('commercial.dashboard.lifetimeEarned')" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial-manager', middleware: 'commercial-manager-only' })

const route = useRoute()
const { t } = useI18n()
const { formatCurrency } = useFormatters()

const overview = ref<any>(null)
const memberName = ref('')
const loading = ref(true)
const loadError = ref(false)

const title = computed(() =>
  memberName.value
    ? t('manager.member.titleNamed', { name: memberName.value })
    : t('manager.member.title'),
)

function formatMix(bps?: number) {
  if (bps == null || bps === 0) return '—'
  return `${(bps / 100).toFixed(1)} %`
}

onMounted(async () => {
  const id = String(route.params.id || '')
  loading.value = true
  loadError.value = false
  try {
    const [ovRes, teamRes]: any[] = await Promise.all([
      $fetch(`/api/commercial-manager/team/${id}/overview`),
      $fetch('/api/commercial-manager/team').catch(() => null),
    ])
    overview.value = ovRes.data ?? ovRes
    const team = teamRes?.data ?? teamRes ?? []
    const m = Array.isArray(team) ? team.find((x: any) => x.userId === id) : null
    memberName.value = m?.fullName || ''
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.pro-link {
  color: var(--pf-vet-accent);
  text-decoration: none;
  font-size: 0.875rem;
}
.pro-link:hover {
  text-decoration: underline;
}
</style>
