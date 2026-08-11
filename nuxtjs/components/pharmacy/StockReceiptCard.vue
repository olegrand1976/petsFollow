<template>
  <ProCard class="pro-mb-lg" data-testid="stock-receipt">
    <h3 class="pro-mb-sm">{{ $t('pharmacy.stock.receiptTitle') }}</h3>
    <div class="stock-form">
      <div>
        <label class="pro-label" for="stock-bl">{{ $t('pharmacy.stock.blNumber') }}</label>
        <input id="stock-bl" v-model="receipt.noteNumber" class="pro-input" data-testid="stock-bl">
      </div>
      <div>
        <label class="pro-label" for="stock-supplier">{{ $t('pharmacy.stock.supplierName') }}</label>
        <input id="stock-supplier" v-model="receipt.supplierName" class="pro-input" data-testid="stock-supplier">
      </div>
      <div v-if="deposits.length > 1">
        <label class="pro-label" for="stock-receipt-deposit">{{ $t('pharmacy.stock.deposit') }}</label>
        <select
          id="stock-receipt-deposit"
          v-model="receipt.depositId"
          class="pro-select"
          data-testid="stock-receipt-deposit"
        >
          <option v-for="d in deposits" :key="d.id" :value="d.id">
            {{ d.name }} ({{ d.code }}){{ d.isDefault ? ` — ${$t('pharmacy.stock.depositDefault')}` : '' }}
          </option>
        </select>
      </div>
      <p
        v-else-if="deposits.length === 1"
        class="pro-hint"
        data-testid="stock-receipt-deposit-hint"
      >
        {{ $t('pharmacy.stock.depositSingleHint', { name: deposits[0].name, code: deposits[0].code }) }}
      </p>
      <div>
        <label class="pro-label" for="stock-med">{{ $t('pharmacy.stock.medication') }}</label>
        <ProCombobox
          input-id="stock-med"
          v-model="selectedMed"
          :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
          :search-fn="searchFn"
          data-testid="stock-med-search"
        />
      </div>
      <div>
        <label class="pro-label" for="stock-lot">{{ $t('pharmacy.stock.lot') }}</label>
        <input id="stock-lot" v-model="receipt.lotNumber" class="pro-input" data-testid="stock-lot">
      </div>
      <div>
        <label class="pro-label" for="stock-exp">{{ $t('pharmacy.stock.expiresOn') }}</label>
        <input id="stock-exp" v-model="receipt.expiresOn" type="date" class="pro-input" data-testid="stock-exp">
      </div>
      <div>
        <label class="pro-label" for="stock-qty">{{ $t('pharmacy.stock.qty') }}</label>
        <input id="stock-qty" v-model.number="receipt.qty" type="number" min="0.001" step="any" class="pro-input" data-testid="stock-qty">
      </div>
      <div class="stock-form__actions">
        <ProButton variant="primary" test-id="stock-receive" :disabled="busy || !canReceive" @click="$emit('receive')">
          {{ $t('pharmacy.stock.receive') }}
        </ProButton>
      </div>
    </div>
    <p v-if="softWarn" class="pro-hint" data-testid="stock-soft-warn">{{ $t('pharmacy.stock.shortDatedWarn') }}</p>
  </ProCard>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import type { PharmacyDeposit } from '~/composables/usePharmacyStockPage'

export type StockReceiptForm = {
  lotNumber: string
  expiresOn: string
  qty: number
  noteNumber: string
  supplierName: string
  depositId: string
}

const receipt = defineModel<StockReceiptForm>('receipt', { required: true })
const selectedMed = defineModel<ProComboboxItem | null>('selectedMed', { required: true })

defineProps<{
  busy: boolean
  canReceive: boolean
  softWarn: boolean
  deposits: PharmacyDeposit[]
  searchFn: (q: string) => Promise<ProComboboxItem[]>
}>()
defineEmits<{ receive: [] }>()
</script>

<style scoped>
.stock-form {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 0.75rem;
  align-items: end;
}
.stock-form__actions { display: flex; align-items: end; gap: 0.5rem; flex-wrap: wrap; }
</style>
