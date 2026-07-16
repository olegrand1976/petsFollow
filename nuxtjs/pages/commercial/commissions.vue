<template>
  <div data-testid="commercial-commissions-page">
    <ProPageHeader
      :title="$t('commercial.commissions.title')"
      :subtitle="$t('commercial.commissions.subtitle')"
    />
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
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'commercial', middleware: 'commercial-only' })

const { formatCurrency } = useFormatters()
const summary = ref<any>(null)

onMounted(async () => {
  const res: any = await $fetch('/api/commercial/commissions')
  summary.value = res.data ?? res
})
</script>
