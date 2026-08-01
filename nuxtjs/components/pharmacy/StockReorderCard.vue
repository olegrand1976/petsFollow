<template>
  <ProCard class="pro-mb-lg" data-testid="stock-reorder">
    <h3 class="pro-mb-sm">{{ $t('pharmacy.stock.reorderTitle') }}</h3>
    <p class="pro-hint pro-mb-sm">{{ $t('pharmacy.stock.reorderHint') }}</p>
    <table class="pro-table" data-testid="stock-reorder-table">
      <thead>
        <tr>
          <th>{{ $t('pharmacy.stock.colMed') }}</th>
          <th>{{ $t('pharmacy.stock.reorderOnHand') }}</th>
          <th>{{ $t('pharmacy.stock.reorderMin') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="a in alerts" :key="a.medicationId">
          <td>
            <strong>{{ a.medicationName }}</strong>
            <div class="pro-hint">{{ a.medicationCnk }}</div>
          </td>
          <td>{{ a.onHandQty }}</td>
          <td>{{ a.minQty }}</td>
        </tr>
      </tbody>
    </table>
    <div class="stock-form">
      <div v-if="suppliers.length">
        <label class="pro-label" for="order-supplier">{{ $t('pharmacy.stock.orderSupplier') }}</label>
        <select
          id="order-supplier"
          class="pro-select"
          data-testid="stock-order-supplier"
          :value="supplierId"
          @change="$emit('update:supplierId', ($event.target as HTMLSelectElement).value)"
        >
          <option value="">{{ $t('pharmacy.stock.orderSupplierPlaceholder') }}</option>
          <option v-for="s in suppliers" :key="s.id" :value="s.id">
            {{ s.name }} — {{ s.email }}
          </option>
        </select>
      </div>
      <template v-else>
        <div>
          <label class="pro-label" for="order-supplier-name">{{ $t('pharmacy.stock.orderSupplierName') }}</label>
          <input
            id="order-supplier-name"
            :value="newSupplierName"
            type="text"
            class="pro-input"
            data-testid="stock-order-supplier-name"
            @input="$emit('update:newSupplierName', ($event.target as HTMLInputElement).value)"
          >
        </div>
        <div>
          <label class="pro-label" for="order-email">{{ $t('pharmacy.stock.orderEmail') }}</label>
          <input
            id="order-email"
            :value="email"
            type="email"
            class="pro-input"
            data-testid="stock-order-email"
            @input="$emit('update:email', ($event.target as HTMLInputElement).value)"
          >
        </div>
      </template>
      <div class="stock-form__actions">
        <ProButton
          variant="primary"
          test-id="stock-order-create-send"
          :disabled="busy || !canSend"
          @click="$emit('send')"
        >
          {{ $t('pharmacy.stock.orderCreateSend') }}
        </ProButton>
      </div>
    </div>
    <p v-if="msg" class="pro-hint" data-testid="stock-order-msg">{{ msg }}</p>
  </ProCard>
</template>

<script setup lang="ts">
import type { PharmacyReorderAlert, PharmacySupplier } from '~/composables/usePharmacyStockPage'

const props = defineProps<{
  alerts: PharmacyReorderAlert[]
  suppliers: PharmacySupplier[]
  supplierId: string
  email: string
  newSupplierName: string
  msg: string
  busy: boolean
}>()
defineEmits<{
  'update:supplierId': [value: string]
  'update:email': [value: string]
  'update:newSupplierName': [value: string]
  send: []
}>()

const canSend = computed(() => {
  if (props.suppliers.length) return Boolean(props.supplierId)
  return Boolean(props.email.trim() && props.newSupplierName.trim())
})
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
