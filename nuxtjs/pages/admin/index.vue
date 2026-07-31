<template>
  <div data-testid="admin-dashboard-page">
    <ProPageHeader
      :title="isDevRole ? $t('admin.dashboard.devTitle') : $t('admin.dashboard.title')"
      :subtitle="isDevRole ? $t('admin.dashboard.devSubtitle') : $t('admin.dashboard.subtitle')"
    />

    <template v-if="isDevRole">
      <div v-if="supportStats" class="pro-grid-kpi" data-testid="admin-dev-support-kpi">
        <ProKpi :value="statusCount('open')" :label="$t('admin.support.statsOpen')" />
        <ProKpi :value="statusCount('in_progress')" :label="$t('admin.support.statsInProgress')" />
        <ProKpi :value="statusCount('resolved')" :label="$t('admin.support.statsResolved')" />
        <ProKpi :value="statusCount('closed')" :label="$t('admin.support.statsClosed')" />
        <ProKpi :value="supportStats.openOlderThan24h ?? 0" :label="$t('admin.support.statsOpen24h')" />
        <ProKpi :value="supportStats.openOlderThan7d ?? 0" :label="$t('admin.support.statsOpen7d')" />
      </div>
    </template>

    <template v-else>
      <div v-if="metrics" class="pro-grid-kpi">
        <ProKpi :value="formatCurrency(metrics.totalRevenueCents)" :label="$t('admin.dashboard.totalRevenue')" />
        <ProKpi :value="formatCurrency(metrics.periodRevenueCents)" :label="$t('admin.dashboard.periodRevenue')" />
        <ProKpi :value="formatCurrency(metrics.mrrCents)" :label="$t('admin.dashboard.mrr')" />
        <ProKpi :value="metrics.userCount" :label="$t('admin.dashboard.registrations')" />
        <ProKpi :value="metrics.petCount" :label="$t('admin.dashboard.pets')" />
        <ProKpi
          :value="`${metrics.conversionRatePercent.toFixed(1)}%`"
          :label="$t('admin.dashboard.conversion')"
        />
        <ProKpi :value="metrics.pendingPayments" :label="$t('admin.dashboard.pendingPayments')" />
        <ProKpi :value="metrics.pastDueCount" :label="$t('admin.dashboard.pastDue')" />
        <ProKpi :value="metrics.commercialCount ?? 0" :label="$t('admin.dashboard.commercials')" />
        <ProKpi :value="metrics.prospectCount ?? 0" :label="$t('admin.dashboard.prospects')" />
        <ProKpi :value="formatCurrency(metrics.addonRevenueCents ?? 0)" :label="$t('admin.dashboard.addonRevenue')" />
        <ProKpi :value="formatCurrency(metrics.commercialCommissionDueCents ?? 0)" :label="$t('admin.dashboard.commercialDue')" />
      </div>
      <div v-if="metrics" class="pro-grid-2 pro-mt-lg">
        <ProCard :title="$t('admin.dashboard.planBreakdown')">
          <div class="pro-bar-chart">
            <div
              v-for="(count, plan) in metrics.planBreakdown"
              :key="plan"
              class="pro-bar-row"
            >
              <span>{{ planLabel(String(plan)) }}</span>
              <div class="pro-bar-track">
                <div
                  class="pro-bar-fill"
                  :style="{ width: barWidth(count, planMax) }"
                />
              </div>
              <span>{{ count }}</span>
            </div>
          </div>
        </ProCard>
        <ProCard :title="$t('admin.dashboard.modeBreakdown')">
          <div class="pro-bar-chart">
            <div
              v-for="(count, mode) in metrics.modeBreakdown"
              :key="mode"
              class="pro-bar-row"
            >
              <span>{{ billingModeLabel(String(mode)) }}</span>
              <div class="pro-bar-track">
                <div
                  class="pro-bar-fill"
                  :style="{ width: barWidth(count, modeMax) }"
                />
              </div>
              <span>{{ count }}</span>
            </div>
          </div>
        </ProCard>
      </div>

      <ProCard
        v-if="stagingSeedEnabled"
        class="pro-mt-lg admin-staging-seed"
        :title="$t('admin.dashboard.stagingSeedTitle')"
        data-testid="admin-staging-seed"
      >
        <p class="admin-staging-seed__warn">{{ $t('admin.dashboard.stagingSeedWarn') }}</p>
        <p class="admin-staging-seed__hint">{{ $t('admin.dashboard.stagingSeedHint', { phrase: confirmPhrase }) }}</p>
        <ProInput
          v-model="confirmInput"
          :label="$t('admin.dashboard.stagingSeedConfirmLabel')"
          autocomplete="off"
          test-id="admin-staging-seed-confirm"
        />
        <p v-if="seedError" class="admin-staging-seed__error" role="alert">{{ seedError }}</p>
        <p v-if="seedOkMsg" class="admin-staging-seed__ok" role="status">{{ seedOkMsg }}</p>
        <div class="admin-staging-seed__actions">
          <ProButton
            variant="secondary"
            :disabled="!canSubmit || seedBusy"
            data-testid="admin-staging-seed-submit"
            @click="runStagingSeed"
          >
            {{ seedBusy ? $t('admin.dashboard.stagingSeedBusy') : $t('admin.dashboard.stagingSeedSubmit') }}
          </ProButton>
        </div>
      </ProCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin-or-dev' })

const CONFIRM_PHRASE = 'RESET STAGING'

const { t } = useI18n()
const { formatCurrency } = useFormatters()
const { planLabel, billingModeLabel } = useCodeLabels()
const { user, fetchUser } = useProUser()
if (!user.value) {
  await fetchUser().catch(() => null)
}
const isDevRole = computed(() => user.value?.role === 'dev')

const metrics = ref<any>(null)
const supportStats = ref<any>(null)
const stagingSeedEnabled = ref(false)
const confirmInput = ref('')
const confirmPhrase = CONFIRM_PHRASE
const seedBusy = ref(false)
const seedError = ref('')
const seedOkMsg = ref('')

const planMax = computed(() =>
  Math.max(...Object.values(metrics.value?.planBreakdown ?? { _: 1 }) as number[], 1),
)
const modeMax = computed(() =>
  Math.max(...Object.values(metrics.value?.modeBreakdown ?? { _: 1 }) as number[], 1),
)

const canSubmit = computed(() => confirmInput.value.trim() === CONFIRM_PHRASE)

function barWidth(count: number, max: number) {
  return `${Math.round((count / max) * 100)}%`
}

function statusCount(status: string) {
  return Number(supportStats.value?.byStatus?.[status] ?? 0)
}

async function runStagingSeed() {
  if (!canSubmit.value || seedBusy.value) return
  if (!confirm(t('admin.dashboard.stagingSeedConfirmDialog'))) return
  seedBusy.value = true
  seedError.value = ''
  seedOkMsg.value = ''
  try {
    const res: any = await $fetch('/api/admin/staging/seed', {
      method: 'POST',
      body: { confirm: CONFIRM_PHRASE },
    })
    const n = Number(res?.data?.notified ?? 0)
    seedOkMsg.value = n > 0
      ? t('admin.dashboard.stagingSeedOk', { n })
      : t('admin.dashboard.stagingSeedOkNoEmail')
    confirmInput.value = ''
    const metricsRes: any = await $fetch('/api/admin/metrics')
    metrics.value = metricsRes.data
  } catch (e: any) {
    const msg = e?.data?.error?.message || e?.data?.message || e?.message
    seedError.value = typeof msg === 'string' && msg ? msg : t('admin.dashboard.stagingSeedFail')
  } finally {
    seedBusy.value = false
  }
}

onMounted(async () => {
  if (isDevRole.value) {
    // Pas d'appel metrics (billing) — support stats only.
    try {
      const statsRes: any = await $fetch('/api/admin/support/stats')
      supportStats.value = statsRes.data ?? statsRes
    } catch {
      supportStats.value = { byStatus: {}, openOlderThan24h: 0, openOlderThan7d: 0 }
    }
    return
  }
  const metricsRes: any = await $fetch('/api/admin/metrics')
  metrics.value = metricsRes.data
  try {
    const seedStatus: any = await $fetch('/api/admin/staging/seed')
    stagingSeedEnabled.value = !!seedStatus?.data?.enabled
  } catch (e: any) {
    stagingSeedEnabled.value = false
    const status = e?.statusCode || e?.status
    if (status && status !== 403 && status !== 404) {
      seedError.value = t('admin.dashboard.stagingSeedStatusFail')
    }
  }
})
</script>

<style scoped>
.admin-staging-seed__warn {
  color: var(--pf-vet-alert);
  font-weight: 600;
  margin: 0 0 0.75rem;
}
.admin-staging-seed__hint {
  margin: 0 0 1rem;
  color: var(--pf-vet-text-muted, #5a6570);
}
.admin-staging-seed__actions {
  margin-top: 1rem;
}
.admin-staging-seed__error {
  color: var(--pf-vet-alert);
  margin: 0.75rem 0 0;
}
.admin-staging-seed__ok {
  color: var(--pf-vet-accent);
  margin: 0.75rem 0 0;
}
</style>
