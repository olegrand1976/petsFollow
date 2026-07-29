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
        <p>{{ $t('pharmacy.legal.waste') }}</p>
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
      :busy="busy"
      @update:email="orderEmail = $event"
      @send="createAndSendOrder"
    />

    <PharmacyStockInventoryCard
      v-if="canWritePharmacy"
      :session="invSession"
      :counts="invCounts"
      :msg="invMsg"
      :busy="busy"
      @start="startInventory"
      @close="closeInventory"
      @cancel="cancelInventory"
      @export="exportInventory"
      @count="onInvCountInput"
    />

    <PharmacyStockReceiptCard
      v-if="canWritePharmacy"
      :receipt="receipt"
      :selected-med="selectedMed"
      :busy="busy"
      :can-receive="canReceive"
      :soft-warn="softWarn"
      :search-fn="searchMedications"
      @update:selected-med="selectedMed = $event"
      @receive="receive"
    />

    <PharmacyStockBatchesTable
      :batches="batches"
      :query="q"
      :busy="busy"
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
:deep(.stock-bands) {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(7.5rem, 1fr));
  gap: 0.5rem;
  margin-bottom: 1rem;
}
:deep(.stock-band) {
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-surface);
  padding: 0.65rem 0.75rem;
  text-align: left;
  cursor: pointer;
}
:deep(.stock-band--active) { outline: 2px solid var(--pf-vet-primary); }
:deep(.stock-band__label) { display: block; font-size: 0.75rem; color: var(--pf-vet-muted); }
:deep(.stock-band__n) { font-size: 1.25rem; }
:deep(.stock-form) {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.75rem;
  align-items: end;
}
:deep(.stock-form__actions) { display: flex; align-items: end; gap: 0.5rem; flex-wrap: wrap; }
:deep(.stock-toolbar) {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}
:deep(.stock-row-actions) {
  display: flex;
  gap: 0.35rem;
  justify-content: flex-end;
  flex-wrap: wrap;
}
</style>
