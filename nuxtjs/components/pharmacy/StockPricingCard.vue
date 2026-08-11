<template>
  <ProCard class="pro-mb-lg" data-testid="stock-pricing">
    <h3 class="pro-mb-sm">{{ $t('pharmacy.stock.pricingTitle') }}</h3>
    <div class="stock-form">
      <div>
        <label class="pro-label">{{ $t('pharmacy.stock.medication') }}</label>
        <ProCombobox
          input-id="stock-pricing-med"
          :model-value="selectedMed"
          :placeholder="$t('pharmacy.medicaments.searchPlaceholder')"
          :search-fn="searchFn"
          data-testid="stock-pricing-med"
          @update:model-value="$emit('update:selectedMed', $event)"
        />
      </div>
      <div>
        <label class="pro-label">{{ $t('pharmacy.stock.pricingPurchase') }}</label>
        <input
          :value="form.purchasePriceCents"
          type="number"
          min="0"
          class="pro-input"
          data-testid="stock-pricing-purchase"
          @input="patchNum('purchasePriceCents', $event)"
        >
      </div>
      <div>
        <label class="pro-label">{{ $t('pharmacy.stock.pricingSell') }}</label>
        <input
          :value="form.sellPriceCents"
          type="number"
          min="0"
          class="pro-input"
          data-testid="stock-pricing-sell"
          @input="patchNum('sellPriceCents', $event)"
        >
      </div>
      <div>
        <label class="pro-label">{{ $t('pharmacy.stock.pricingVat') }}</label>
        <input
          :value="form.vatPercent"
          type="number"
          min="0"
          step="0.01"
          class="pro-input"
          data-testid="stock-pricing-vat"
          @input="patchNum('vatPercent', $event)"
        >
      </div>
      <div>
        <label class="pro-label">{{ $t('pharmacy.stock.pricingMinQty') }}</label>
        <input
          :value="form.minQty"
          type="number"
          min="0"
          step="any"
          class="pro-input"
          data-testid="stock-pricing-min"
          @input="patchNum('minQty', $event)"
        >
      </div>
      <ProButton
        variant="primary"
        test-id="stock-pricing-save"
        :disabled="busy || !selectedMed?.id"
        @click="$emit('save')"
      >
        {{ $t('pharmacy.stock.pricingSave') }}
      </ProButton>
      <p v-if="msg" class="pro-hint" data-testid="stock-pricing-msg">{{ msg }}</p>
    </div>
  </ProCard>
</template>

<script setup lang="ts">
import type { ProComboboxItem } from '~/components/pro/ProCombobox.vue'
import type { PharmacyPricingForm } from '~/composables/usePharmacyStockPage'

const props = defineProps<{
  selectedMed: ProComboboxItem | null
  form: PharmacyPricingForm
  busy: boolean
  msg: string
  searchFn: (q: string) => Promise<ProComboboxItem[]>
}>()
const emit = defineEmits<{
  'update:selectedMed': [value: ProComboboxItem | null]
  'update:form': [value: PharmacyPricingForm]
  save: []
}>()

function patchNum(key: keyof PharmacyPricingForm, e: Event) {
  const n = Number((e.target as HTMLInputElement).value)
  emit('update:form', { ...props.form, [key]: Number.isFinite(n) ? n : 0 })
}
</script>

<style scoped>
.stock-form {
  display: grid;
  gap: 0.65rem;
  grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
  align-items: end;
}
</style>
