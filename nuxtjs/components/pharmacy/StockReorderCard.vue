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
      <div class="stock-form__actions">
        <ProButton
          variant="primary"
          test-id="stock-order-create-send"
          :disabled="busy || !email"
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
import type { PharmacyReorderAlert } from '~/composables/usePharmacyStockPage'

defineProps<{
  alerts: PharmacyReorderAlert[]
  email: string
  msg: string
  busy: boolean
}>()
defineEmits<{
  'update:email': [value: string]
  send: []
}>()
</script>
