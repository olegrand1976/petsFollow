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
      <div>
        <label class="pro-label" for="stock-med">{{ $t('pharmacy.stock.medication') }}</label>
        <ProCombobox
          input-id="stock-med"
          :model-value="selectedMed"
          :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
          :search-fn="searchFn"
          data-testid="stock-med-search"
          @update:model-value="$emit('update:selectedMed', $event)"
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

defineProps<{
  receipt: {
    lotNumber: string
    expiresOn: string
    qty: number
    noteNumber: string
    supplierName: string
  }
  selectedMed: ProComboboxItem | null
  busy: boolean
  canReceive: boolean
  softWarn: boolean
  searchFn: (q: string) => Promise<ProComboboxItem[]>
}>()
defineEmits<{
  receive: []
  'update:selectedMed': [value: ProComboboxItem | null]
}>()
</script>
