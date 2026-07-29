<template>
  <div data-testid="stock-page">
    <ProPageHeader
      :title="$t('pharmacy.stock.title')"
      :subtitle="$t('pharmacy.stock.subtitle')"
    >
      <template #actions>
        <ProBadge variant="warning" data-testid="stock-page-dev-badge">{{ $t('nav.tagDev') }}</ProBadge>
        <ProButton variant="secondary" test-id="stock-export" @click="exportCsv">
          {{ $t('pharmacy.stock.export') }}
        </ProButton>
      </template>
    </ProPageHeader>

    <details class="pharmacy-legal" data-testid="pharmacy-legal">
      <summary>{{ $t('pharmacy.legal.title') }}</summary>
      <div class="pharmacy-legal__body">
        <p>{{ $t('pharmacy.legal.fefo') }}</p>
        <p>{{ $t('pharmacy.legal.expiry') }}</p>
        <p>{{ $t('pharmacy.legal.quarantine') }}</p>
        <p>{{ $t('pharmacy.legal.waste') }}</p>
        <p>{{ $t('pharmacy.legal.careVsStock') }}</p>
      </div>
    </details>

    <p v-if="error" class="pro-alert pro-mb-md" data-testid="stock-error">{{ error }}</p>

    <PharmacyStockBands
      :cards="bandCards"
      :active="bandFilter"
      :summary="summary"
      @select="setBand"
    />

    <PharmacyStockReorderCard
      v-if="canWritePharmacy && reorderAlerts.length"
      :alerts="reorderAlerts"
      :email="orderEmail"
      :msg="orderMsg"
      :busy="busyOrder"
      @update:email="orderEmail = $event"
      @send="createAndSendOrder"
    />

    <PharmacyStockInventoryCard
      v-if="canWritePharmacy"
      :session="invSession"
      :counts="invCounts"
      :msg="invMsg"
      :busy="busyInventory"
      @start="startInventory"
      @close="closeInventory"
      @cancel="cancelInventory"
      @export="exportInventory"
      @count="onInvCountInput"
    />

    <PharmacyStockReceiptCard
      v-if="canWritePharmacy"
      v-model:receipt="receipt"
      v-model:selected-med="selectedMed"
      :busy="busyReceipt"
      :can-receive="canReceive"
      :soft-warn="softWarn"
      :search-fn="searchMedications"
      @receive="receive"
    />

    <PharmacyStockBatchesTable
      :batches="batches"
      :query="q"
      :busy="busy || busyBatchAction"
      :can-write="canWritePharmacy"
      :band-variant="bandVariant"
      :band-label="bandLabel"
      @update:query="q = $event"
      @search="loadBatches"
      @refresh="refresh"
      @quarantine="quarantine"
      @waste="waste"
    />
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: ['vet-only', 'practice-perm'], practicePerm: 'pharmacy.read' })

const {
  canWritePharmacy,
  busy,
  busyReceipt,
  busyOrder,
  busyInventory,
  busyBatchAction,
  error,
  softWarn,
  bandFilter,
  q,
  selectedMed,
  receipt,
  batches,
  reorderAlerts,
  orderEmail,
  orderMsg,
  invSession,
  invCounts,
  invMsg,
  summary,
  bandCards,
  canReceive,
  bandVariant,
  bandLabel,
  setBand,
  searchMedications,
  loadBatches,
  startInventory,
  onInvCountInput,
  closeInventory,
  cancelInventory,
  exportInventory,
  refresh,
  createAndSendOrder,
  receive,
  quarantine,
  waste,
  exportCsv,
} = usePharmacyStockPage()

onMounted(refresh)
</script>

<style scoped>
.pharmacy-legal {
  margin: 0 0 1rem;
  padding: 0.65rem 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
}
.pharmacy-legal summary {
  cursor: pointer;
  font-weight: 600;
}
.pharmacy-legal__body {
  margin-top: 0.5rem;
  display: grid;
  gap: 0.35rem;
}
.pharmacy-legal__body p { margin: 0; font-size: 0.9rem; color: var(--pf-vet-muted); }
</style>
