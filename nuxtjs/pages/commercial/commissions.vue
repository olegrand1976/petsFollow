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
      <ProKpi :value="`${(summary.rateBps / 100).toFixed(0)}%`" :label="$t('commercial.commissions.rate')" />
      <ProKpi :value="formatCurrency(summary.monthEarnedCents)" :label="$t('commercial.commissions.month')" />
      <ProKpi :value="formatCurrency(summary.lifetimeEarnedCents)" :label="$t('commercial.commissions.lifetime')" />
      <ProKpi :value="formatCurrency(summary.subscriptionCommissionCents)" :label="$t('commercial.commissions.subscriptions')" />
      <ProKpi :value="formatCurrency(summary.addonCommissionCents)" :label="$t('commercial.commissions.addons')" />
    </div>
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

onMounted(async () => {
  const [commRes, profileRes]: any[] = await Promise.all([
    $fetch('/api/commercial/commissions'),
    $fetch('/api/commercial/me/payout-profile').catch(() => null),
  ])
  summary.value = commRes.data ?? commRes
  const profile = profileRes?.data ?? profileRes
  hasIban.value = Boolean(profile?.iban)
})
</script>

<style scoped>
.pro-banner {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  margin-bottom: 1rem;
  padding: 0.75rem 1rem;
  border: 1px solid var(--pf-vet-border);
  background: var(--pf-vet-surface);
}
.pro-mt-lg {
  margin-top: 1.25rem;
}
</style>
