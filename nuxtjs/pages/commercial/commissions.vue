<template>
  <div data-testid="commercial-commissions-page">
    <ProPageHeader
      :title="$t('commercial.commissions.title')"
      :subtitle="$t('commercial.commissions.subtitle')"
    />
    <div v-if="!hasIban" class="pro-banner">
      <span>{{ $t('commercial.commissions.ibanBanner') }}</span>
      <NuxtLink to="/commercial/settings">{{ $t('commercial.commissions.ibanBannerLink') }}</NuxtLink>
    </div>
    <div v-if="summary" class="pro-grid-kpi">
      <ProKpi :value="`${(summary.rateBps / 100).toFixed(0)}%`" :label="$t('commercial.commissions.rateHeart')" />
      <ProKpi :value="formatCurrency(summary.monthEarnedCents)" :label="$t('commercial.commissions.month')" />
      <ProKpi :value="formatCurrency(summary.lifetimeEarnedCents)" :label="$t('commercial.commissions.lifetime')" />
      <ProKpi :value="formatCurrency(summary.subscriptionCommissionCents)" :label="$t('commercial.commissions.subscriptions')" />
      <ProKpi :value="formatCurrency(summary.addonCommissionCents)" :label="$t('commercial.commissions.addons')" />
    </div>

    <ProCard :title="$t('commercial.commissions.gainsTitle')" class="pro-mt-lg">
      <div class="pf-plan-compare">
        <div
          v-for="row in summary?.planRates || []"
          :key="row.code"
          class="pf-plan-compare__item"
          :class="{ 'pf-plan-compare__item--rec': row.recommended }"
        >
          <strong>{{ $t(`commissionSheet.plans.${row.code}`) }}</strong>
          <span>{{ formatPct(row.commercialRateBps) }} · {{ formatCurrency(row.commercialCents) }}</span>
          <ProBadge v-if="row.recommended" variant="success">{{ $t('commissionSheet.recommended') }}</ProBadge>
        </div>
      </div>
      <div class="pf-bonus-row" data-testid="commercial-bonus-cards">
        <ProCard v-for="b in commercialBonuses" :key="b.code" class="pf-bonus-card" :data-testid="`bonus-card-${b.code}`">
          <strong>{{ $t(`commissionSheet.bonusTitles.${b.code}`) }}</strong>
          <p>{{ formatCurrency(b.amountCents) }}</p>
          <p class="text-muted">{{ $t(`commissionSheet.bonusHints.${b.code}`) }}</p>
          <p v-if="b.code === 'commercial_ramp' && (b.vetFullName || b.vetEmail)" class="text-muted">
            {{ $t('commercial.commissions.bonusVet', { vet: b.vetFullName || b.vetEmail }) }}
          </p>
          <p v-if="b.code === 'commercial_mix' && b.periodYm" class="text-muted">
            {{ $t('commercial.commissions.bonusPeriod', { period: b.periodYm }) }}
          </p>
          <ProBadge :variant="bonusBadgeVariant(b.status)">
            {{ $t(`commissionSheet.status.${b.status || 'available'}`) }}
            <template v-if="b.progress != null && b.target">
              — {{ b.progress }}/{{ b.target }}<template v-if="b.code === 'commercial_mix'"> %</template>
            </template>
          </ProBadge>
        </ProCard>
      </div>
    </ProCard>

    <ProCard :title="$t('commercial.commissions.sheetTitle')" class="pro-mt-lg">
      <ProCommissionSheet
        audience="commercial"
        :plan-rates="summary?.planRates || []"
        :addon-rates="summary?.addonRates || []"
        :bonuses="summary?.bonuses || []"
      />
      <NuxtLink to="/commercial/pitch" class="pro-hint-link">{{ $t('commercial.commissions.pitchLink') }}</NuxtLink>
    </ProCard>

    <ProCard class="pro-mt-lg">
      <ProTable :empty="!summary?.recentLedger?.length" :empty-title="$t('commercial.commissions.empty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.commissions.date') }}</th>
            <th>{{ $t('commercial.commissions.type') }}</th>
            <th>{{ $t('commercial.commissions.vet') }}</th>
            <th>{{ $t('commercial.commissions.client') }}</th>
            <th>{{ $t('commercial.commissions.base') }}</th>
            <th>{{ $t('commercial.commissions.amount') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in summary?.recentLedger || []" :key="row.id">
            <td>{{ row.accruedAt?.substring?.(0, 10) || row.accruedAt }}</td>
            <td><ProBadge variant="neutral">{{ row.sourceType }}</ProBadge></td>
            <td>{{ row.vetEmail }}</td>
            <td>{{ row.clientEmail }}</td>
            <td>{{ formatCurrency(row.baseAmountCents) }}</td>
            <td>{{ formatCurrency(row.commissionCents) }}</td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
    <ProCard :title="$t('commercial.commissions.payoutsTitle')" class="pro-mt-lg">
      <ProTable :empty="!(summary?.payoutHistory || []).length" :empty-title="$t('commercial.commissions.payoutsEmpty')">
        <thead>
          <tr>
            <th>{{ $t('commercial.commissions.colPeriod') }}</th>
            <th>{{ $t('commercial.commissions.colAmount') }}</th>
            <th>{{ $t('commercial.commissions.colStatus') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in summary?.payoutHistory || []" :key="p.periodYm">
            <td>{{ p.periodYm }}</td>
            <td>{{ formatCurrency(p.amountCents) }}</td>
            <td><ProBadge :variant="p.runStatus === 'paid' ? 'success' : 'warning'">{{ p.runStatus }}</ProBadge></td>
          </tr>
        </tbody>
      </ProTable>
    </ProCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { formatCurrency } = useFormatters()
const summary = ref<any>(null)
const hasIban = ref(true)

const commercialBonuses = computed(() =>
  (summary.value?.bonuses || []).filter((b: any) => b.audience === 'commercial'),
)

function formatPct(bps: number) {
  return `${((bps || 0) / 100).toFixed(0)} %`
}

function bonusBadgeVariant(status?: string): 'success' | 'warning' | 'neutral' {
  if (status === 'paid' || status === 'earned') return 'success'
  if (status === 'in_progress') return 'warning'
  return 'neutral'
}

onMounted(async () => {
  const [commRes, profileRes]: any[] = await Promise.all([
    $fetch('/api/commercial/commissions'),
    $fetch('/api/commercial/me/payout-profile').catch(() => null),
  ])
  summary.value = commRes.data ?? commRes
  const profile = profileRes?.data ?? profileRes
  if (profile && (profile.iban != null || profile.payoutIban != null)) {
    hasIban.value = Boolean(profile.iban || profile.payoutIban)
  }
})
</script>

<style scoped>
.pro-grid-kpi {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.75rem;
}
.pro-mt-lg { margin-top: 1.25rem; }
.pro-banner {
  margin-bottom: 1rem;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  background: var(--pf-vet-surface);
  border: 1px solid var(--pf-vet-border);
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}
.pf-plan-compare {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
}
.pf-plan-compare__item {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  padding: 0.75rem;
  display: grid;
  gap: 0.35rem;
}
.pf-plan-compare__item--rec { border-color: var(--pf-vet-accent); }
.pf-bonus-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}
.pf-bonus-card p { margin: 0.25rem 0; }
.pro-hint-link {
  display: inline-block;
  margin-top: 0.75rem;
  color: var(--pf-vet-accent);
}
@media (max-width: 900px) {
  .pf-plan-compare { grid-template-columns: 1fr; }
}
</style>
