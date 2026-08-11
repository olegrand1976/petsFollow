<template>
  <ProCard data-testid="stock-batches">
    <div class="stock-toolbar">
      <input
        :value="query"
        class="pro-input"
        :placeholder="$t('pharmacy.stock.searchPlaceholder')"
        data-testid="stock-search"
        @input="$emit('update:query', ($event.target as HTMLInputElement).value)"
        @keyup.enter="$emit('search')"
      >
      <select
        v-if="deposits.length > 1"
        class="pro-select"
        data-testid="stock-deposit-filter"
        :value="depositFilter"
        @change="$emit('update:depositFilter', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">{{ $t('pharmacy.stock.depositFilterAll') }}</option>
        <option v-for="d in deposits" :key="d.id" :value="d.id">{{ d.name }} ({{ d.code }})</option>
      </select>
      <ProButton variant="secondary" test-id="stock-refresh" :disabled="busy" @click="$emit('refresh')">
        {{ $t('pharmacy.stock.refresh') }}
      </ProButton>
    </div>
    <div v-if="!batches.length" class="pro-empty" data-testid="stock-empty">
      {{ $t('pharmacy.stock.empty') }}
    </div>
    <table v-else class="pro-table" data-testid="stock-table">
      <thead>
        <tr>
          <th>{{ $t('pharmacy.stock.colMed') }}</th>
          <th>{{ $t('pharmacy.stock.lot') }}</th>
          <th>{{ $t('pharmacy.stock.expiresOn') }}</th>
          <th>{{ $t('pharmacy.stock.qty') }}</th>
          <th>{{ $t('pharmacy.stock.band') }}</th>
          <th />
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in batches" :key="row.id" :data-testid="`stock-row-${row.id}`">
          <td>
            <strong>{{ row.medicationName }}</strong>
            <div class="pro-hint">{{ row.medicationCnk }} · {{ row.depositCode }}</div>
          </td>
          <td>{{ row.lotNumber }}</td>
          <td>{{ row.expiresOn }}</td>
          <td>{{ row.qtyOnHand }} {{ row.unit }}</td>
          <td>
            <ProBadge :variant="bandVariant(row.expiryBand)">{{ bandLabel(row.expiryBand) }}</ProBadge>
          </td>
          <td class="stock-row-actions">
            <ProButton
              v-if="canWrite && row.status === 'active'"
              variant="secondary"
              :test-id="`stock-quarantine-${row.id}`"
              :disabled="busy"
              @click="$emit('quarantine', row.id)"
            >
              {{ $t('pharmacy.stock.quarantine') }}
            </ProButton>
            <ProButton
              v-if="canWrite && row.status === 'active'"
              variant="secondary"
              :test-id="`stock-adjust-${row.id}`"
              :disabled="busy"
              @click="beginAdjust(row.id)"
            >
              {{ $t('pharmacy.stock.adjust') }}
            </ProButton>
            <ProButton
              v-if="canWrite"
              variant="ghost"
              :test-id="`stock-waste-${row.id}`"
              :disabled="busy"
              @click="beginWaste(row.id)"
            >
              {{ $t('pharmacy.stock.waste') }}
            </ProButton>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="pendingWasteId" class="stock-dialog" data-testid="stock-waste-dialog">
      <h4>{{ $t('pharmacy.stock.wasteReason') }}</h4>
      <select v-model="wasteReason" class="pro-select" data-testid="stock-waste-reason">
        <option value="expired">{{ $t('pharmacy.stock.wasteReasonExpired') }}</option>
        <option value="supplier_return">{{ $t('pharmacy.stock.wasteReasonSupplierReturn') }}</option>
        <option value="destruction">{{ $t('pharmacy.stock.wasteReasonDestruction') }}</option>
      </select>
      <div class="stock-dialog__actions">
        <ProButton variant="ghost" test-id="stock-waste-cancel" @click="pendingWasteId = null">
          {{ $t('pharmacy.stock.cancelAction') }}
        </ProButton>
        <ProButton variant="primary" test-id="stock-waste-confirm" :disabled="busy" @click="confirmWaste">
          {{ $t('pharmacy.stock.wasteSubmit') }}
        </ProButton>
      </div>
    </div>

    <div v-if="pendingAdjustId" class="stock-dialog" data-testid="stock-adjust-dialog">
      <h4>{{ $t('pharmacy.stock.adjust') }}</h4>
      <label class="pro-label">{{ $t('pharmacy.stock.adjustPrompt') }}</label>
      <input v-model.number="adjustDelta" type="number" step="any" class="pro-input" data-testid="stock-adjust-delta">
      <div class="stock-dialog__actions">
        <ProButton variant="ghost" test-id="stock-adjust-cancel" @click="pendingAdjustId = null">
          {{ $t('pharmacy.stock.cancelAction') }}
        </ProButton>
        <ProButton variant="primary" test-id="stock-adjust-confirm" :disabled="busy || !adjustDelta" @click="confirmAdjust">
          {{ $t('pharmacy.stock.adjustConfirm') }}
        </ProButton>
      </div>
    </div>
  </ProCard>
</template>

<script setup lang="ts">
import type { PharmacyBatchRow, PharmacyDeposit } from '~/composables/usePharmacyStockPage'
import type { PharmacyWasteReason } from '~/utils/pharmacy-stock'

defineProps<{
  batches: PharmacyBatchRow[]
  deposits: PharmacyDeposit[]
  depositFilter: string
  query: string
  busy: boolean
  canWrite: boolean
  bandVariant: (band: string) => 'success' | 'warning' | 'danger' | 'neutral'
  bandLabel: (band: string) => string
}>()
const emit = defineEmits<{
  'update:query': [value: string]
  'update:depositFilter': [value: string]
  search: []
  refresh: []
  quarantine: [id: string]
  waste: [id: string, reason: PharmacyWasteReason]
  adjust: [id: string, delta: number]
}>()

const pendingWasteId = ref<string | null>(null)
const wasteReason = ref<PharmacyWasteReason>('expired')
const pendingAdjustId = ref<string | null>(null)
const adjustDelta = ref(0)

function beginWaste(id: string) {
  pendingAdjustId.value = null
  pendingWasteId.value = id
  wasteReason.value = 'expired'
}

function confirmWaste() {
  if (!pendingWasteId.value) return
  emit('waste', pendingWasteId.value, wasteReason.value)
  pendingWasteId.value = null
}

function beginAdjust(id: string) {
  pendingWasteId.value = null
  pendingAdjustId.value = id
  adjustDelta.value = 0
}

function confirmAdjust() {
  if (!pendingAdjustId.value || !adjustDelta.value) return
  emit('adjust', pendingAdjustId.value, adjustDelta.value)
  pendingAdjustId.value = null
}
</script>

<style scoped>
.stock-toolbar {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1rem;
}
.stock-row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  justify-content: flex-end;
}
.stock-dialog {
  margin-top: 1rem;
  padding: 0.85rem;
  border: 1px solid var(--pf-vet-border);
  border-radius: 8px;
  background: var(--pf-vet-bg);
  display: grid;
  gap: 0.65rem;
  max-width: 24rem;
}
.stock-dialog__actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}
</style>
